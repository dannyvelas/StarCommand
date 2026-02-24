package main

import (
	"fmt"

	"github.com/dannyvelas/conflux"
	"github.com/dannyvelas/starcommand/internal/app"
	"github.com/spf13/cobra"
)

func terraformApplyCmd(configMux *conflux.ConfigMux, preflight bool) *cobra.Command {
	terraformApplyCmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply the terraform project",
		RunE:  terraformApplyCLI(configMux, preflight),
	}

	return terraformApplyCmd
}

func terraformApplyCLI(configMux *conflux.ConfigMux, preflight bool) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		diagnostics, err := app.TerraformApply(ctx, configMux, preflight)
		if err != nil {
			return err
		}

		if len(diagnostics) > 0 {
			fmt.Printf("%s\n", app.DiagnosticsToTable(diagnostics))
		}

		return nil
	}
}
