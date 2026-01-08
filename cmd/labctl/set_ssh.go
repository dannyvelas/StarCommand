package main

import (
	"github.com/spf13/cobra"
)

func setSSHCmd() *cobra.Command {
	setSSHCmd := &cobra.Command{
		Use:   "ssh <host-name>",
		Short: "Update the `~/.ssh/config` file to connect to a given host",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			host := args[0]
		},
	}

	return setSSHCmd
}
