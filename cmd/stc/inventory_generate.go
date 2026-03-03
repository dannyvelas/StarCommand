package main

import (
	"fmt"

	"github.com/dannyvelas/starcommand/config"
	"github.com/dannyvelas/starcommand/internal/app"
	"github.com/spf13/cobra"
)

func inventoryGenerateCmd(c *config.Config, preflight bool) *cobra.Command {
	var host string

	command := &cobra.Command{
		Use:   "generate",
		Short: "Generate the Ansible inventory file for all hosts, or a single host",
		RunE:  inventoryGenerateCLI(c, preflight, &host),
	}

	command.Flags().StringVar(&host, "host", "", "Limit generation to a single host")

	return command
}

func inventoryGenerateCLI(c *config.Config, preflight bool, host *string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		diagnostics, err := app.InventoryGenerate(ctx, c, host, preflight)
		if err != nil {
			return err
		}

		if len(diagnostics) > 0 {
			fmt.Printf("%s\n", app.DiagnosticsToTable(diagnostics))
		}
		return nil
	}
}
