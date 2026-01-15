package main

import (
	"encoding/json"
	"errors"
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
			configMux := config.NewConfigMux(
				hostName,
				verbose,
				config.WithFileReader(),
				config.WithEnvReader(),
				config.WithBitwardenSecretReader(),
			)

			// TODO: change this to be dynamic
			proxmoxConfig := hosts.NewProxmox()
			diagnostics, err := config.Unmarshal(configMux, proxmoxConfig)
			if err != nil && !errors.Is(err, config.ErrInvalidFields) {
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
				os.Exit(1)
			}

			if dryRun {
				fmt.Fprintf(os.Stderr, "Configs for %s:\n%s\n", hostName, config.DiagnosticsToTable(diagnostics))
				return
			} else if errors.Is(err, config.ErrInvalidFields) {
				fmt.Fprintf(os.Stderr, "%v for %s:\n%s\n", config.ErrInvalidFields, hostName, config.DiagnosticsToTable(diagnostics))
				return
			}

			bytes, err := json.MarshalIndent(proxmoxConfig, "", "    ")
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
