package main

import (
	"github.com/dannyvelas/conflux"
	"github.com/spf13/cobra"
)

func sshCmd(configMux *conflux.ConfigMux, preflight bool) *cobra.Command {
	sshCmd := &cobra.Command{
		Use:   "ssh",
		Short: "ssh-related utilities",
	}

	sshCmd.AddCommand(sshAddCmd(configMux, preflight))

	return sshCmd
}
