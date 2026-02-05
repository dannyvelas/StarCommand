package handlers

import (
	"context"
	"fmt"
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
	terraformProxmoxConfig, ok := config.(*terraformProxmoxConfig)
	if !ok {
		return nil, fmt.Errorf("internal type error converting config to terraform proxmox config. found: %T", config)
	}

	return executeTerraformFlow(ctx, terraformProxmoxConfig, h.terraformFilePath, terraformProxmoxConfig.TerraformVersionConstraint)
}
