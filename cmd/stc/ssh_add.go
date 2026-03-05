package main

import (
	"fmt"

	"github.com/dannyvelas/starcommand/config"
	"github.com/dannyvelas/starcommand/internal/app"
	"github.com/spf13/cobra"
)

func sshAddCmd(c *config.Config) *cobra.Command {
	sshAddCmd := &cobra.Command{
		Use:   "add <host>",
		Short: "Add a host to ~/.ssh/config",
		Args:  cobra.ExactArgs(1),
		RunE:  sshAddCLI(c),
	}

	return sshAddCmd
}

func sshAddCLI(c *config.Config) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		hostAlias := args[0]
		diagnostics, err := app.SSHAdd(ctx, c, hostAlias)
		if err != nil {
			return err
		}

		if len(diagnostics) > 0 {
			fmt.Printf("%s\n", app.DiagnosticsToTable(diagnostics))
		}

		return nil
	}
}
