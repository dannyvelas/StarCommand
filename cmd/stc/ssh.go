package main

import (
	"github.com/dannyvelas/starcommand/internal/models"
	"github.com/spf13/cobra"
)

func sshCmd(c *models.Config) *cobra.Command {
	sshCmd := &cobra.Command{
		Use:   "ssh",
		Short: "ssh-related utilities",
	}

	sshCmd.AddCommand(sshAddCmd(c))

	return sshCmd
}
