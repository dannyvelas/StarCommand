package main

import (
	"fmt"

	"github.com/dannyvelas/conflux"
	"github.com/dannyvelas/starcommand/internal/app"
	"github.com/spf13/cobra"
)

func sshAddCmd(configMux *conflux.ConfigMux, preflight bool) *cobra.Command {
	sshAddCmd := &cobra.Command{
		Use:   "add <host>",
		Short: "Add a host to ~/.ssh/config",
		Args:  cobra.ExactArgs(1),
		RunE:  sshAddCLI(configMux, preflight),
	}

	return sshAddCmd
}

func sshAddCLI(configMux *conflux.ConfigMux, preflight bool) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		hostAlias := args[0]
		diagnostics, err := app.SSHAdd(ctx, configMux, hostAlias, preflight)
		if err != nil {
			return err
		}

		if len(diagnostics) > 0 {
			fmt.Printf("%s\n", app.DiagnosticsToTable(diagnostics))
		}

		return nil
	}
}
