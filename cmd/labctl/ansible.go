package main

import (
	"github.com/spf13/cobra"
)

func getCmd() *cobra.Command {
	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Display information for a resource",
	}

	getCmd.AddCommand(getConfigCmd())

	return getCmd
}
