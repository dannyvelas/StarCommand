package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/dannyvelas/homelab/internal/config"
	"github.com/spf13/cobra"
)

func getConfigCmd(verbose bool) *cobra.Command {
	var dryRun bool

	getConfigCmd := &cobra.Command{
		Use:   "config <host-name>",
		Short: "Generate a JSON object of configuration values for a given host",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			host := args[0]
			fullConfig := config.NewFullConfig(host, verbose)

			if dryRun {
				validation, err := fullConfig.DryRun(host, verbose)
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s\n", err.Error())
					os.Exit(1)
				}

				fmt.Printf("Config Requirements for %s:\n%s", host, validation)
				return
			}

			config, err := fullConfig.ReadValidated()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
				os.Exit(1)
			}

			bytes, err := json.MarshalIndent(config, "", "    ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "error marshalling to JSON: %s", err.Error())
				os.Exit(1)
			}

			fmt.Println(string(bytes))
		},
	}

	getConfigCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print report of found/missing keys for <host-name>")

	return getConfigCmd
}
