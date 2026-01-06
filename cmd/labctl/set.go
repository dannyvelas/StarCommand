package main

import "github.com/spf13/cobra"

func setCmd() *cobra.Command {
	setCmd := &cobra.Command{
		Use:   "set",
		Short: "Display one or many resources",
	}

	setCmd.AddCommand(setSSHCmd())

	return setCmd
}
