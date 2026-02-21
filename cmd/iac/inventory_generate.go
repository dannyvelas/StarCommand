package main

import (
	"fmt"

	"github.com/dannyvelas/conflux"
	"github.com/dannyvelas/homelab/internal/app"
	"github.com/spf13/cobra"
)

func inventoryGenerateCmd(configMux *conflux.ConfigMux, preflight bool) *cobra.Command {
	command := &cobra.Command{
		Use:   "generate",
		Short: "Generate the Ansible inventory file for all hosts, or a single host",
		Args:  cobra.ExactArgs(1),
		RunE:  inventoryGenerateCLI(configMux, preflight),
	}

	return command
}

func inventoryGenerateCLI(configMux *conflux.ConfigMux, preflight bool) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		diagnostics, err := app.InventoryGenerate(ctx, configMux, preflight)
		if err != nil {
			return err
		}

		if len(diagnostics) > 0 {
			fmt.Printf("%s\n", app.DiagnosticsToTable(diagnostics))
		}
		return nil
	}
}
