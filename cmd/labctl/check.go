package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/dannyvelas/conflux"
	"github.com/dannyvelas/homelab/internal/app"
	"github.com/dannyvelas/homelab/internal/helpers"
	"github.com/spf13/cobra"
)

func checkCmd() *cobra.Command {
	checkCmd := &cobra.Command{
		Use:   "check <host-alias> target1 [targets]",
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

			fmt.Printf("Configs needed for host(%s):\n%s\n", hostAlias, app.DiagnosticsToTable(diagnostics))
		},
	}

	return checkCmd
}

func toTargets(args []string) ([]app.Target, error) {
	targets := make([]app.Target, 0)
	for _, arg := range args {
		target, err := toTarget(arg)
		if err != nil {
			return nil, err
		}
		targets = append(targets, target)
	}
	return targets, nil
}

func toTarget(arg string) (app.Target, error) {
	split := strings.Split(arg, ":")

	resource, err := parseResource(arg, split)
	if err != nil {
		return app.Target{}, err
	}

	action, err := parseAction(split)
	if err != nil {
		return app.Target{}, err
	}

	return app.Target{Resource: resource, Action: action}, nil
}

func parseResource(arg string, split []string) (app.Resource, error) {
	first, rest, err := shift(split)
	if err != nil {
		return "", fmt.Errorf("error: invalid target argument: %s", arg)
	}

	switch first {
	case "ansible":
		return parseAnsibleResource(rest)
	case "ssh":
		return app.SSHResource, nil
	case "terraform":
		return app.TerraformResource, nil
	default:
		return "", fmt.Errorf("error: unrecognized resource: %s", first)
	}
}

func parseAnsibleResource(split []string) (app.Resource, error) {
	first, _, err := shift(split)
	if err != nil {
		return "", fmt.Errorf("error: expecting ansible sub-command")
	}

	switch first {
	case "playbook":
		return app.AnsiblePlaybookResource, nil
	case "inventory":
		return app.AnsibleInventoryResource, nil
	default:
		return "", fmt.Errorf("error: unrecognized ansible sub-command: %s", first)
	}
}

func parseAction(split []string) (app.Action, error) {
	first, _, err := shift(split)
	if err != nil {
		return "", fmt.Errorf("error: expecting action after resource")
	}

	action, err := app.StringToAction(first)
	if err != nil {
		return "", fmt.Errorf("error: unrecognized action: %s", first)
	}

	return action, nil
}

func shift(s []string) (string, []string, error) {
	if len(s) < 1 {
		return "", s, fmt.Errorf("empty slice")
	}

	return s[0], s[1:], nil
}
