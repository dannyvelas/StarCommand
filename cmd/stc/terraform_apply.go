package main

import (
	"github.com/dannyvelas/starcommand/internal/app"
	"github.com/dannyvelas/starcommand/internal/models"
	"github.com/spf13/cobra"
)

func terraformApplyCmd(c *models.Config) *cobra.Command {
	terraformApplyCmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply the terraform project",
		RunE:  terraformApplyCLI(c),
	}

	return terraformApplyCmd
}

func terraformApplyCLI(c *models.Config) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		return app.TerraformApply(ctx, c)
	}
}
