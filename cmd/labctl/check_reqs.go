package main

import (
	"errors"
	"fmt"
	"maps"
	"os"

	"github.com/dannyvelas/conflux"
	"github.com/dannyvelas/homelab/internal/helpers"
	"github.com/dannyvelas/homelab/internal/models"
	"github.com/spf13/cobra"
)

func checkReqsCmd() *cobra.Command {
	checkReqsCmd := &cobra.Command{
		Use:   "reqs <host-alias>",
		Short: "Print a diagnostic report of all the configs that were found/missing for a given resource",
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
			configDiagnostics, err := conflux.Unmarshal(configMux, proxmoxConfig)
			if err != nil && !errors.Is(err, conflux.ErrInvalidFields) {
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
				os.Exit(1)
			}

			sshHost := models.NewSSHHost(hostAlias)
			sshDiagnostics, err := conflux.Unmarshal(configMux, sshHost)
			if err != nil && !errors.Is(err, conflux.ErrInvalidFields) {
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
				os.Exit(1)
			}

			// merge maps
			maps.Copy(configDiagnostics, sshDiagnostics)

			fmt.Printf("Configs for %s:\n%s\n", hostAlias, conflux.DiagnosticsToTable(configDiagnostics))
		},
	}

	return checkReqsCmd
}
