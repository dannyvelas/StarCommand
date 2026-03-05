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

func (h terraformHandler) execute(ctx context.Context, config *terraformConfig) error {
	return executeTerraformFlow(ctx, config, h.terraformFilePath, config.TerraformVersionConstraint)
}
