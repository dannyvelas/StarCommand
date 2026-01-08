package main

import (
	"github.com/spf13/cobra"
)

func getCmd(verbose bool) *cobra.Command {
	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Display one or many resources",
	}

	getCmd.AddCommand(getConfigCmd(verbose))

	return getCmd
}
