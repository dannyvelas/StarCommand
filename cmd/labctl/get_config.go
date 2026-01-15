package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/dannyvelas/homelab/internal/config"
	"github.com/dannyvelas/homelab/internal/hosts"
	"github.com/spf13/cobra"
)

func getConfigCmd(verbose bool) *cobra.Command {
	var dryRun bool

	getConfigCmd := &cobra.Command{
		Use:   "config <host-name>",
		Short: "Generate a JSON object of configuration values for a given host",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			hostName := args[0]
			fullConfigReader := config.NewFullConfigReader(hostName, verbose)

			if dryRun {
				//validation, err := fullConfigReader.DryRun()
				//if err != nil {
				//	fmt.Fprintf(os.Stderr, "%s\n", err.Error())
				//	os.Exit(1)
				//}

				// fmt.Printf("Config Requirements for %s:\n%s", hostName, validation)
				return
			}

			proxmoxConfig := hosts.NewProxmox()
			c, err := config.Unmarshal(fullConfigReader, proxmoxConfig)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
				os.Exit(1)
			}

			bytes, err := json.MarshalIndent(c, "", "    ")
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
