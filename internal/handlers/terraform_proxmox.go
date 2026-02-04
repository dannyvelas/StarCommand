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
	"github.com/hashicorp/hc-install/fs"
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

	if err := h.upsertTerraformConstraint(terraformProxmoxConfig.TerraformVersionConstraint); errors.Is(err, errAlreadyExists) {
		diagnosticKey := fmt.Sprintf("Setting terraform version in %s", h.terraformFilePath)
		diagnostics[diagnosticKey] = fmt.Sprintf("skipping: %v", errAlreadyExists)
	} else if err != nil {
		return diagnostics, fmt.Errorf("error creating token for terraform user: %v", err)
	}

	execPath, doneFn, err := h.locateTerraform(ctx, terraformProxmoxConfig.TerraformVersionConstraint)
	defer doneFn()
	if err != nil {
		return diagnostics, fmt.Errorf("error locating terraform executable: %v", err)
	}

	if err := h.applyTerraform(ctx, terraformProxmoxConfig, execPath); err != nil {
		return diagnostics, fmt.Errorf("error applying terraform project: %v", err)
	}

	return diagnostics, nil
}

func (h TerraformProxmoxHandler) upsertTerraformConstraint(desiredConstraint string) error {
	src, err := os.ReadFile(h.terraformFilePath)
	if err != nil {
		return fmt.Errorf("error reading terraform file at %s: %v", h.terraformFilePath, err)
	}

	newFileByes, err := transformTerraformVersion(src, h.terraformFilePath, desiredConstraint)
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

func transformTerraformVersion(src []byte, filePath string, constraint string) ([]byte, error) {
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
			if t.Type == hclsyntax.TokenQuotedLit && string(t.Bytes) == constraint {
				return nil, errAlreadyExists
			}
		}
	}

	tfBody.SetAttributeValue("required_version", cty.StringVal(constraint))

	return f.Bytes(), nil
}

func (h TerraformProxmoxHandler) locateTerraform(ctx context.Context, desiredConstraint string) (string, func(), error) {
	installer := install.NewInstaller()
	doneFn := func() { _ = installer.Remove(ctx) }

	constraints := version.MustConstraints(version.NewConstraint(desiredConstraint))
	execPath, err := installer.Ensure(ctx, []src.Source{
		&fs.Version{
			Product:     product.Terraform,
			Constraints: constraints,
		},
		&releases.LatestVersion{
			Product:     product.Terraform,
			Constraints: constraints,
		},
	})
	if err != nil {
		return "", doneFn, fmt.Errorf("error locating/installing version %s: %v", desiredConstraint, err)
	}

	return execPath, doneFn, nil
}

func (h TerraformProxmoxHandler) applyTerraform(ctx context.Context, config *terraformProxmoxConfig, execPath string) error {
	tf, err := tfexec.NewTerraform(filepath.Dir(h.terraformFilePath), execPath)
	if err != nil {
		return fmt.Errorf("error running NewTerraform: %s", err)
	}

	tf.SetStdout(os.Stdout)
	tf.SetStderr(os.Stderr)

	tmpFile, err := os.CreateTemp("", "labctl-vars-*.tfvars.json")
	if err != nil {
		return fmt.Errorf("error creating temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	if err := json.NewEncoder(tmpFile).Encode(config); err != nil {
		return fmt.Errorf("error writing config to tmp file: %v", err)
	}

	if err := tf.Apply(ctx, tfexec.VarFile(tmpFile.Name())); err != nil {
		return fmt.Errorf("error applying terraform file: %v", err)
	}

	return nil
}
