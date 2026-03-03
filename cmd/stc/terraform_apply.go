package main

import (
	"fmt"

	"github.com/dannyvelas/starcommand/config"
	"github.com/dannyvelas/starcommand/internal/app"
	"github.com/spf13/cobra"
)

func terraformApplyCmd(c *config.Config, preflight bool) *cobra.Command {
	terraformApplyCmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply the terraform project",
		RunE:  terraformApplyCLI(c, preflight),
	}

	return terraformApplyCmd
}

func terraformApplyCLI(c *config.Config, preflight bool) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		diagnostics, err := app.TerraformApply(ctx, c, preflight)
		if err != nil {
			return err
		}

		if len(diagnostics) > 0 {
			fmt.Printf("%s\n", app.DiagnosticsToTable(diagnostics))
		}

		return nil
	}
}
