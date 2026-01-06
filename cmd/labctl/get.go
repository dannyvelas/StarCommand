package main

import (
	"github.com/dannyvelas/homelab/internal/env"
	"github.com/spf13/cobra"
)

func getCmd(env env.Env, verbose bool) *cobra.Command {
	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Display one or many resources",
	}

	getCmd.AddCommand(getConfigCmd(env, verbose))

	return getCmd
}
