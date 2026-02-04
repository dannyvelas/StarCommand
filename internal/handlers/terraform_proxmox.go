package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-version"
	install "github.com/hashicorp/hc-install"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/hc-install/src"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/zclconf/go-cty/cty"
)

type terraformBlock struct {
	RequiredVersion string   `hcl:"required_version,optional"`
	Remainder       hcl.Body `hcl:",remain"`
}

var _ Handler = TerraformProxmoxHandler{}

type TerraformProxmoxHandler struct {
	terraformFilePath string
}

func NewTerraformProxmoxHandler() TerraformProxmoxHandler {
	return TerraformProxmoxHandler{
		terraformFilePath: "./terraform/global/firewall.tf",
	}
}

func (h TerraformProxmoxHandler) GetConfig(_ string) any {
	return newTerraformProxmoxConfig()
}

func (h TerraformProxmoxHandler) Execute(config any, hostAlias string) (map[string]string, error) {
	ctx := context.Background()
	diagnostics := make(map[string]string)

	terraformProxmoxConfig, ok := config.(*terraformProxmoxConfig)
	if !ok {
		return diagnostics, fmt.Errorf("internal type error converting config to terraform proxmox config. found: %T", config)
	}

	diagnosticKey := fmt.Sprintf("Setting terraform version in %s", h.terraformFilePath)
	if err := h.upsertTerraformVersion(terraformProxmoxConfig.TerraformVersion); errors.Is(err, errAlreadyExists) {
		diagnostics[diagnosticKey] = fmt.Sprintf("skipping: %v", errAlreadyExists)
	} else if err != nil {
		return diagnostics, fmt.Errorf("error creating token for terraform user: %v", err)
	}

	execPath, err := h.locateTerraform(terraformProxmoxConfig.TerraformVersion)
	if err != nil {
		return diagnostics, fmt.Errorf("error locating terraform executable: %v", err)
	}

	if err := h.applyTerraform(ctx, terraformProxmoxConfig, execPath); err != nil {
		return diagnostics, fmt.Errorf("error running terraform playbook: %v", err)
	}

	return diagnostics, nil
}

func (h TerraformProxmoxHandler) upsertTerraformVersion(desiredVersion string) error {
	currentTerraformVersion, err := h.getTerraformVersion()
	if err != nil && !errors.Is(err, errNotFound) {
		return fmt.Errorf("error getting terraform version: %v", err)
	}

	// if found same version, no need to update
	if err == nil && currentTerraformVersion == desiredVersion {
		return errAlreadyExists
	}

	// if found different version, or if did not find terraform version, set new version
	if err := h.setTerraformVersion(desiredVersion); err != nil {
		return fmt.Errorf("error setting terraform version: %v", err)
	}

	return nil
}

func (h TerraformProxmoxHandler) getTerraformVersion() (string, error) {
	parser := hclparse.NewParser()

	file, diags := parser.ParseHCLFile(h.terraformFilePath)
	if diags.HasErrors() {
		return "", fmt.Errorf("error parsing HCL file: %v", diags)
	}

	schema := &hcl.BodySchema{Blocks: []hcl.BlockHeaderSchema{{Type: "terraform"}}}
	content, _, diags := file.Body.PartialContent(schema)
	if diags.HasErrors() {
		return "", fmt.Errorf("error getting partial content from hcl file %s: %v", h.terraformFilePath, diags)
	}

	for _, block := range content.Blocks {
		if block.Type != "terraform" {
			continue
		}

		var tf terraformBlock
		decodeDiags := gohcl.DecodeBody(block.Body, nil, &tf)
		if decodeDiags.HasErrors() {
			return "", fmt.Errorf("error decoding terraform block: %v", decodeDiags)
		}

		if tf.RequiredVersion == "" {
			continue
		}

		return tf.RequiredVersion, nil
	}

	return "", fmt.Errorf("terraform block %w", errNotFound)
}

func (h TerraformProxmoxHandler) setTerraformVersion(desiredVersion string) error {
	src, err := os.ReadFile(h.terraformFilePath)
	if err != nil {
		return fmt.Errorf("error reading terraform file at %s: %v", h.terraformFilePath, err)
	}

	f, diags := hclwrite.ParseConfig(src, h.terraformFilePath, hcl.InitialPos)
	if diags.HasErrors() {
		return fmt.Errorf("error parsing HCL: %v", diags)
	}

	rootBody := f.Body()

	tfBlock := rootBody.FirstMatchingBlock("terraform", nil)
	if tfBlock == nil {
		tfBlock = rootBody.AppendNewBlock("terraform", nil)
	}

	tfBody := tfBlock.Body()
	tfBody.SetAttributeValue("required_version", cty.StringVal(desiredVersion))

	if err := os.WriteFile(h.terraformFilePath, f.Bytes(), 0o644); err != nil {
		return fmt.Errorf("error writing file: %v", err)
	}

	return nil
}

func (h TerraformProxmoxHandler) locateTerraform(desiredVersion string) (string, error) {
	ctx := context.Background()
	installer := install.NewInstaller()
	defer installer.Remove(ctx)

	v := version.Must(version.NewVersion(desiredVersion))
	execPath, err := installer.Install(ctx, []src.Installable{
		&releases.ExactVersion{
			Product: product.Terraform,
			Version: v,
		},
	})
	if err != nil {
		return "", fmt.Errorf("error locating version %s: %v", desiredVersion, err)
	}

	return execPath, nil
}

func (h TerraformProxmoxHandler) applyTerraform(ctx context.Context, config *terraformProxmoxConfig, execPath string) error {
	tf, err := tfexec.NewTerraform(filepath.Dir(h.terraformFilePath), execPath)
	if err != nil {
		return fmt.Errorf("error running NewTerraform: %s", err)
	}

	tmpFile, err := os.CreateTemp("", "labctl-vars-*.tfvars.json")
	if err != nil {
		return fmt.Errorf("error creating temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if err := json.NewEncoder(tmpFile).Encode(config); err != nil {
		return fmt.Errorf("error writing config to tmp file: %v", err)
	}

	if err := tf.Apply(ctx, tfexec.VarFile(tmpFile.Name())); err != nil {
		return fmt.Errorf("error applying terraform file: %v", err)
	}

	return nil
}
