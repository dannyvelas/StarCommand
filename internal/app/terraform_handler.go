package app

import (
	"context"
)

type terraformHandler struct {
	terraformFilePath string
}

func newTerraformHandler(terraformFilePath string) terraformHandler {
	return terraformHandler{
		terraformFilePath: terraformFilePath,
	}
}

func (h terraformHandler) execute(ctx context.Context, config *terraformConfig) (map[string]string, error) {
	return nil, executeTerraformFlow(ctx, config, h.terraformFilePath, config.TerraformVersionConstraint)
}
