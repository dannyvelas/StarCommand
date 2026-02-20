package app

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/dannyvelas/conflux"
	"github.com/dannyvelas/homelab/internal/handlers"
)

func AnsibleRun(ctx context.Context, configMux *conflux.ConfigMux, playbook string, preflight bool) (map[string]string, error) {
	return nil, nil
}

func SSHAdd(ctx context.Context, configMux *conflux.ConfigMux, hostAlias string, preflight bool) (map[string]string, error) {
	sshConfig := handlers.NewSSHHost(hostAlias)
	diagnostics, err := conflux.Unmarshal(configMux, sshConfig)
	if preflight {
		return diagnostics, nil
	} else if errors.Is(err, conflux.ErrInvalidFields) {
		return diagnostics, fmt.Errorf("error getting config for adding ssh config for host %s:\n%s", hostAlias, conflux.DiagnosticsToTable(diagnostics))
	} else if err != nil {
		return nil, fmt.Errorf("internal error unmarshalling ssh config to struct: %v", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("error getting user home dir: %v", err)
	}

	sshHandler := handlers.NewSSHHandler(homeDir)
	handlerDiagnostics, err := sshHandler.Execute(sshConfig, hostAlias)
	if err != nil {
		return nil, fmt.Errorf("error executing command: %v", err)
	}

	return handlerDiagnostics, nil
}

func TerraformApply(ctx context.Context, configMux *conflux.ConfigMux, preflight bool) (map[string]string, error) {
	return nil, nil
}

func execute(ctx context.Context, configMux *conflux.ConfigMux, resource resource, action action, hostAlias string, dryRun bool) (map[string]string, error) {
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

	handlerDiagnostics, err := rule.handler.Execute(ctx, configStruct, hostAlias)
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
