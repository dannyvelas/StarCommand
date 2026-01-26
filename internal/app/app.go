package app

import (
	"fmt"
	"maps"

	"github.com/dannyvelas/conflux"
	"github.com/dannyvelas/homelab/internal/models"
)

type rule struct {
	Name    string
	Match   func(resource models.Resource, action models.Action, hostAlias string) bool
	Execute func(configMux *conflux.ConfigMux, hostAlias string, dryRun bool) (map[string]string, error)
}

type Target struct {
	Resource models.Resource
	Action   models.Action
}

func AnsibleRun(configMux *conflux.ConfigMux, hostAlias string) (map[string]string, error) {
	return execute(configMux, models.AnsiblePlaybookResource, models.RunAction, hostAlias, false)
}

func SSHAdd(configMux *conflux.ConfigMux, hostAlias string) (map[string]string, error) {
	return execute(configMux, models.SSHResource, models.RunAction, hostAlias, false)
}

func Check(configMux *conflux.ConfigMux, hostAlias string, targets []Target) (map[string]string, error) {
	allDiagnostics := make(map[string]string)
	for _, target := range targets {
		diagnostics, err := execute(configMux, target.Resource, target.Action, hostAlias, true)
		if err != nil {
			return nil, fmt.Errorf("error checking resource(%s), action(%s), and alias(%s)", target.Resource, target.Action, hostAlias)
		}
		maps.Copy(allDiagnostics, diagnostics)
	}

	return allDiagnostics, nil
}

func execute(configMux *conflux.ConfigMux, resource models.Resource, action models.Action, hostAlias string, dryRun bool) (map[string]string, error) {
	for _, rule := range registry {
		if rule.Match(resource, action, hostAlias) {
			return rule.Execute(configMux, hostAlias, dryRun)
		}
	}

	return nil, fmt.Errorf("error: unsupported combination of resource(%s), action(%s), and hostAlias(%s)", resource, action, hostAlias)
}

var registry = []rule{
	{
		Name: "ansible run proxmox",
		Match: func(resource models.Resource, action models.Action, host string) bool {
			return resource == models.AnsiblePlaybookResource && action == models.RunAction && host == "proxmox"
		},
		Execute: func(configMux *conflux.ConfigMux, host string, dryRun bool) (map[string]string, error) {
			if dryRun {
				return map[string]string{"terraform": "Plan: Create VM"}, nil
			}
			// ... Actual Logic
			return map[string]string{}, nil
		},
	},
	{
		Name: "Default Ansible",
		Match: func(resource models.Resource, action models.Action, host string) bool {
			return resource == models.AnsiblePlaybookResource && action == models.RunAction
		},
		Execute: func(configMux *conflux.ConfigMux, host string, dryRun bool) (map[string]string, error) {
			// ... Logic
			return map[string]string{}, nil
		},
	},
}
