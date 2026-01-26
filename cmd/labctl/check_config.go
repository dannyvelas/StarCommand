package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/dannyvelas/conflux"
	"github.com/dannyvelas/homelab/internal/app"
	"github.com/dannyvelas/homelab/internal/helpers"
	"github.com/dannyvelas/homelab/internal/models"
	"github.com/spf13/cobra"
)

func checkConfigCmd() *cobra.Command {
	checkConfigCmd := &cobra.Command{
		Use: "config <host-alias>",
		// TODO: fix
		// ValidArgs: handlers.GetSupportedHostAliases(),
		Short: "Print a diagnostic report of all the configs that were found/missing for a given resource",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			hostAlias := args[0]
			configMux := conflux.NewConfigMux(
				conflux.WithYAMLFileReader(helpers.FallbackFile, conflux.WithPath(helpers.GetConfigPath(hostAlias))),
				conflux.WithEnvReader(),
				conflux.WithBitwardenSecretReader(),
			)

			targets, err := toTargets(args[1:])
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
				os.Exit(1)
			}

			diagnostics, err := app.Check(configMux, hostAlias, targets)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
				os.Exit(1)
			}

			fmt.Printf("Configs needed for %s:\n%s\n", hostAlias, app.DiagnosticsToTable(diagnostics))
		},
	}

	return checkConfigCmd
}

func toTargets(slice []string) ([]app.Target, error) {
	targets := make([]app.Target, 0)
	malformedArgs := make([]string, 0)
	for _, el := range slice {
		split := strings.Split(el, ":")
		amtArgs := len(split)
		if amtArgs < 2 {
			malformedArgs = append(malformedArgs, el)
			continue
		}

		resourceString := split[amtArgs-2]
		resource, err := models.StringToResource(resourceString)
		if err != nil {
			malformedArgs = append(malformedArgs, el)
			continue
		}

		actionString := split[amtArgs-1]
		action, err := models.StringToAction(actionString)
		if err != nil {
			malformedArgs = append(malformedArgs, el)
			continue
		}

		targets = append(targets, app.Target{Resource: resource, Action: action})
	}

	if len(malformedArgs) > 0 {
		return nil, fmt.Errorf("error: malformed args: %v", malformedArgs)
	}

	return targets, nil
}
