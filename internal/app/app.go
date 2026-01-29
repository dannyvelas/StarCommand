package app

import (
	"maps"

	"github.com/dannyvelas/conflux"
)

func AnsibleRun(configMux *conflux.ConfigMux, hostAlias string) (map[string]string, error) {
	return execute(configMux, AnsiblePlaybookResource, RunAction, hostAlias, false)
}

func SSHAdd(configMux *conflux.ConfigMux, hostAlias string) (map[string]string, error) {
	return execute(configMux, SSHResource, AddAction, hostAlias, false)
}

func Check(configMux *conflux.ConfigMux, hostAlias string, targets []Target) (map[string]string, error) {
	allDiagnostics := make(map[string]string)
	for _, target := range targets {
		diagnostics, err := execute(configMux, target.resource, target.action, hostAlias, true)
		if err != nil {
			return nil, err
		}
		maps.Copy(allDiagnostics, diagnostics)
	}

	return allDiagnostics, nil
}
