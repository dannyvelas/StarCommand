package main

import (
	"github.com/dannyvelas/starcommand/config"
	"github.com/spf13/cobra"
)

func sshCmd(c *config.Config, preflight bool) *cobra.Command {
	sshCmd := &cobra.Command{
		Use:   "ssh",
		Short: "ssh-related utilities",
	}

	sshCmd.AddCommand(sshAddCmd(c, preflight))

	return sshCmd
}
