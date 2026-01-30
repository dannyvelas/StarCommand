package app

import (
	"errors"
	"fmt"

	"github.com/dannyvelas/conflux"
	"github.com/dannyvelas/homelab/internal/handlers"
)

var errUnsupportedCombination = errors.New("error: unsupported combination")

type rule struct {
	Name    string
	Match   func(resource resource, action action, hostAlias string) bool
	handler handlers.Handler
}

var registry = []rule{
	{
		Name: "ansible run proxmox",
		Match: func(resource resource, action action, hostAlias string) bool {
			return resource == ansiblePlaybookResource && action == runAction && hostAlias == "proxmox"
		},
		handler: handlers.NewAnsibleProxmoxHandler(),
	},
	{
		Name: "ssh add <any-host-alias>",
		Match: func(resource resource, action action, hostAlias string) bool {
			return resource == sshResource && action == addAction
		},
		handler: handlers.NewSSHHandler(),
	},
	{
		Name: "terraform apply proxmox",
		Match: func(resource resource, action action, hostAlias string) bool {
			return resource == terraformResource && action == applyAction && hostAlias == "proxmox"
		},
		handler: handlers.NewTerraformProxmoxHandler(),
	},
}

func execute(configMux *conflux.ConfigMux, resource resource, action action, hostAlias string, dryRun bool) (map[string]string, error) {
	rule, err := findRule(resource, action, hostAlias)
	if err != nil {
		return nil, err
	}

	configStruct := rule.handler.GetConfig(hostAlias)
	diagnostics, err := conflux.Unmarshal(configMux, configStruct)
	if dryRun {
		return diagnostics, nil
	} else if errors.Is(err, conflux.ErrInvalidFields) {
		return diagnostics, fmt.Errorf("error getting config for running %s on %s host:\n%s", resource, hostAlias, conflux.DiagnosticsToTable(diagnostics))
	} else if err != nil {
		return nil, fmt.Errorf("internal error unmarshalling config to struct for %s on %s: %v", resource, hostAlias, err)
	}

	handlerDiagnostics, err := rule.handler.Execute(configStruct, hostAlias)
	if err != nil {
		return nil, fmt.Errorf("error executing command: %v", err)
	}

	return handlerDiagnostics, nil
}

func findRule(resource resource, action action, hostAlias string) (rule, error) {
	for _, rule := range registry {
		if rule.Match(resource, action, hostAlias) {
			return rule, nil
		}
	}

	return rule{}, fmt.Errorf("%w of resource(%s), action(%s), and hostAlias(%s)", errUnsupportedCombination, resource, action, hostAlias)
}
