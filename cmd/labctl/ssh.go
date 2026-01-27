package main

import "github.com/spf13/cobra"

func sshCmd() *cobra.Command {
	sshCmd := &cobra.Command{
		Use:   "ssh",
		Short: "Execute ssh commands",
	}

	sshCmd.AddCommand(sshAddCmd())

	return sshCmd
}
