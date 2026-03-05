package main

import (
	"github.com/dannyvelas/starcommand/config"
	"github.com/dannyvelas/starcommand/internal/app"
	"github.com/spf13/cobra"
)

func terraformApplyCmd(c *config.Config) *cobra.Command {
	terraformApplyCmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply the terraform project",
		RunE:  terraformApplyCLI(c),
	}

	return terraformApplyCmd
}

func terraformApplyCLI(c *config.Config) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		return app.TerraformApply(ctx, c)
	}
}
