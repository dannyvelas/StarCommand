package main

import "github.com/spf13/cobra"

func setCmd() *cobra.Command {
	setCmd := &cobra.Command{
		Use:   "set",
		Short: "Create a resource",
	}

	setCmd.AddCommand(setSSHCmd())

	return setCmd
}
