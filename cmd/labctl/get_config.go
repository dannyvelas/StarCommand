package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/dannyvelas/conflux"
	"github.com/dannyvelas/homelab/internal/helpers"
	"github.com/dannyvelas/homelab/internal/models"
	"github.com/spf13/cobra"
)

func getConfigCmd() *cobra.Command {
	var dryRun bool

	getConfigCmd := &cobra.Command{
		Use:   "config <host-alias>",
		Short: "Generate a JSON object of configuration values for a given host",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			hostAlias := args[0]

			configMux := conflux.NewConfigMux(
				conflux.WithYAMLFileReader(helpers.FallbackFile, conflux.WithPath(helpers.GetConfigPath(hostAlias))),
				conflux.WithEnvReader(),
				conflux.WithBitwardenSecretReader(),
			)

			// TODO: change this to be dynamic
			proxmoxConfig := models.NewProxmox()
			diagnostics, err := conflux.Unmarshal(configMux, proxmoxConfig)
			if err != nil && !errors.Is(err, conflux.ErrInvalidFields) {
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
				os.Exit(1)
			}

			if dryRun {
				fmt.Fprintf(os.Stderr, "Configs for %s:\n%s\n", hostAlias, conflux.DiagnosticsToTable(diagnostics))
				return
			} else if errors.Is(err, conflux.ErrInvalidFields) {
				fmt.Fprintf(os.Stderr, "%v for %s:\n%s\n", conflux.ErrInvalidFields, hostAlias, conflux.DiagnosticsToTable(diagnostics))
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

	getConfigCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print report of found/missing keys for <host-alias>")

	return getConfigCmd
}
