package app

import (
	"fmt"
	"maps"

	"github.com/dannyvelas/conflux"
)

type Target struct {
	Resource Resource
	Action   Action
}

func AnsibleRun(configMux *conflux.ConfigMux, hostAlias string) (map[string]string, error) {
	return execute(configMux, AnsiblePlaybookResource, RunAction, hostAlias, false)
}

func SSHAdd(configMux *conflux.ConfigMux, hostAlias string) (map[string]string, error) {
	return execute(configMux, SSHResource, AddAction, hostAlias, false)
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
