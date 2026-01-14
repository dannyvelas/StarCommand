package main

import (
	"github.com/dannyvelas/homelab/internal/config"
	"github.com/spf13/cobra"
)

func setSSHCmd(verbose bool) *cobra.Command {
	setSSHCmd := &cobra.Command{
		Use:   "ssh <host-name>",
		Short: "Update the `~/.ssh/config` file to connect to a given host",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			hostName := args[0]
			fullConfigReader := config.NewFullConfigReader(hostName, verbose)
			sshSetter := newSSHSetter(hostName, fullConfigReader)
		},
	}

	return setSSHCmd
}
