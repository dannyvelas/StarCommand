package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/hashicorp/go-version"
	install "github.com/hashicorp/hc-install"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/hc-install/src"
	"github.com/hashicorp/terraform-exec/tfexec"
)

var _ Handler = TerraformProxmoxHandler{}

type TerraformProxmoxHandler struct{}

func NewTerraformProxmoxHandler() TerraformProxmoxHandler {
	return TerraformProxmoxHandler{}
}

func (h TerraformProxmoxHandler) GetConfig(_ string) any {
	return newTerraformProxmoxConfig()
}

func (h TerraformProxmoxHandler) Execute(config any, hostAlias string) (map[string]string, error) {
	diagnostics := make(map[string]string)

	terraformProxmoxConfig, ok := config.(*terraformProxmoxConfig)
	if !ok {
		return diagnostics, fmt.Errorf("internal type error converting config to terraform proxmox config. found: %T", config)
	}

	if err := h.applyTerraform(terraformProxmoxConfig); err != nil {
		return diagnostics, fmt.Errorf("error running terraform playbook: %v", err)
	}

	return diagnostics, nil
}

func (h TerraformProxmoxHandler) applyTerraform(config *terraformProxmoxConfig) error {
	ctx := context.Background()
	installer := install.NewInstaller()
	defer installer.Remove(ctx)

	v := version.Must(version.NewVersion(config.TerraformVersion))
	execPath, err := installer.Install(ctx, []src.Installable{
		&releases.ExactVersion{
			Product: product.Terraform,
			Version: v,
		},
	})
	if err != nil {
		return fmt.Errorf("error locating version %s: %v", config.TerraformVersion, err)
	}

	tf, err := tfexec.NewTerraform("./terraform/global", execPath)
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
