package main

import (
	"github.com/dannyvelas/starcommand/internal/app"
	"github.com/dannyvelas/starcommand/internal/config"
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
		hostAlias := args[0]
		return app.SSHAdd(cmd.Context(), c, hostAlias)
	}
}
