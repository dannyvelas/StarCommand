package handlers

import (
	"context"
	"fmt"
)

var _ Handler = TerraformPlexHandler{}

type TerraformPlexHandler struct {
	terraformFilePath string
}

func NewTerraformPlexHandler(terraformFilePath string) TerraformPlexHandler {
	return TerraformPlexHandler{
		terraformFilePath: terraformFilePath,
	}
}

func (h TerraformPlexHandler) GetConfig(_ string) any {
	return newTerraformPlexConfig()
}

func (h TerraformPlexHandler) Execute(ctx context.Context, config any, hostAlias string) (map[string]string, error) {
	terraformPlexConfig, ok := config.(*terraformPlexConfig)
	if !ok {
		return nil, fmt.Errorf("internal type error converting config to terraform plex config. found: %T", config)
	}

	return nil, executeTerraformFlow(ctx, terraformPlexConfig, h.terraformFilePath, terraformPlexConfig.TerraformVersionConstraint)
}
