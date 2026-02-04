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
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/zclconf/go-cty/cty"
)

var _ Handler = TerraformProxmoxHandler{}

type TerraformProxmoxHandler struct {
	terraformFilePath string
}

func NewTerraformProxmoxHandler(terraformFilePath string) TerraformProxmoxHandler {
	return TerraformProxmoxHandler{
		terraformFilePath: terraformFilePath,
	}
}

func (h TerraformProxmoxHandler) GetConfig(_ string) any {
	return newTerraformProxmoxConfig()
}

func (h TerraformProxmoxHandler) Execute(ctx context.Context, config any, hostAlias string) (map[string]string, error) {
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

	execPath, err := h.locateTerraform(ctx, terraformProxmoxConfig.TerraformVersion)
	if err != nil {
		return diagnostics, fmt.Errorf("error locating terraform executable: %v", err)
	}

	if err := h.applyTerraform(ctx, terraformProxmoxConfig, execPath); err != nil {
		return diagnostics, fmt.Errorf("error running terraform playbook: %v", err)
	}

	return diagnostics, nil
}

func (h TerraformProxmoxHandler) upsertTerraformVersion(desiredVersion string) error {
	src, err := os.ReadFile(h.terraformFilePath)
	if err != nil {
		return fmt.Errorf("error reading terraform file at %s: %v", h.terraformFilePath, err)
	}

	newFileByes, err := transformTerraformVersion(src, h.terraformFilePath, desiredVersion)
	if err != nil && !errors.Is(err, errAlreadyExists) {
		return fmt.Errorf("error transforming terraform file to have new version: %v", err)
	}

	if errors.Is(err, errAlreadyExists) {
		return errAlreadyExists
	}

	if err := os.WriteFile(h.terraformFilePath, newFileByes, 0o644); err != nil {
		return fmt.Errorf("error writing file: %v", err)
	}

	return nil
}

func transformTerraformVersion(src []byte, filePath string, version string) ([]byte, error) {
	f, diags := hclwrite.ParseConfig(src, filePath, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to parse HCL: %s", diags)
	}

	rootBody := f.Body()
	tfBlock := rootBody.FirstMatchingBlock("terraform", nil)

	if tfBlock == nil {
		tfBlock = rootBody.AppendNewBlock("terraform", nil)
	}

	tfBody := tfBlock.Body()
	attr := tfBody.GetAttribute("required_version")
	if attr != nil {
		tokens := attr.Expr().BuildTokens(nil)
		for _, t := range tokens {
			if t.Type == hclsyntax.TokenQuotedLit && string(t.Bytes) == version {
				return nil, errAlreadyExists
			}
		}
	}

	tfBody.SetAttributeValue("required_version", cty.StringVal(version))

	return f.Bytes(), nil
}

func (h TerraformProxmoxHandler) locateTerraform(ctx context.Context, desiredVersion string) (string, error) {
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
