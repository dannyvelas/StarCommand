package main

import "github.com/spf13/cobra"

func setCmd(verbose bool) *cobra.Command {
	setCmd := &cobra.Command{
		Use:   "set",
		Short: "Display one or many resources",
	}

	setCmd.AddCommand(setSSHCmd(verbose))

	return setCmd
}
