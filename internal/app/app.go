package app

import (
	"context"
	"maps"

	"github.com/dannyvelas/conflux"
)

func AnsibleRun(ctx context.Context, configMux *conflux.ConfigMux, hostAlias string) (map[string]string, error) {
	return execute(ctx, configMux, ansiblePlaybookResource, runAction, hostAlias, false)
}

func SSHAdd(ctx context.Context, configMux *conflux.ConfigMux, hostAlias string) (map[string]string, error) {
	return execute(ctx, configMux, sshResource, addAction, hostAlias, false)
}

func TerraformApply(ctx context.Context, configMux *conflux.ConfigMux, hostAlias string) (map[string]string, error) {
	return execute(ctx, configMux, terraformResource, applyAction, hostAlias, false)
}

func Check(ctx context.Context, configMux *conflux.ConfigMux, hostAlias string, targetArgs []string) (map[string]string, error) {
	targets, err := toTargets(targetArgs)
	if err != nil {
		return nil, err
	}

	allDiagnostics := make(map[string]string)
	for _, target := range targets {
		diagnostics, err := execute(ctx, configMux, target.resource, target.action, hostAlias, true)
		if err != nil {
			return nil, err
		}
		maps.Copy(allDiagnostics, diagnostics)
	}

	return allDiagnostics, nil
}
