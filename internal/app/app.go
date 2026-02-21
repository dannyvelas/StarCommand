package app

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/dannyvelas/conflux"
)

func Setup(ctx context.Context, configMux *conflux.ConfigMux, hostAliases []string, preflight bool) (map[string]string, error) {
	return nil, nil
}

func InventoryGenerate(ctx context.Context, configMux *conflux.ConfigMux, preflight bool) (map[string]string, error) {
	return nil, nil
}

func AnsibleRun(ctx context.Context, configMux *conflux.ConfigMux, playbook string, preflight bool) (map[string]string, error) {
	ansibleHandler := newAnsibleHandler()

	playbookConfig, err := ansibleHandler.getConfig(playbook)
	if err != nil {
		return nil, err
	}

	diagnostics, err := conflux.Unmarshal(configMux, playbookConfig)
	if preflight {
		return diagnostics, nil
	} else if errors.Is(err, conflux.ErrInvalidFields) {
		return diagnostics, fmt.Errorf("error getting config for running %s playbook:\n%s", playbook, conflux.DiagnosticsToTable(diagnostics))
	} else if err != nil {
		return nil, fmt.Errorf("internal error unmarshalling config to struct for %s playbook: %v", playbook, err)
	}

	handlerDiagnostics, err := ansibleHandler.execute(playbookConfig)
	if err != nil {
		return nil, fmt.Errorf("error executing command: %v", err)
	}

	return handlerDiagnostics, nil
}

func SSHAdd(ctx context.Context, configMux *conflux.ConfigMux, hostAlias string, preflight bool) (map[string]string, error) {
	sshConfig := newSSHHost(hostAlias)
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

	sshHandler := newSSHHandler(homeDir)
	handlerDiagnostics, err := sshHandler.execute(sshConfig, hostAlias)
	if err != nil {
		return nil, fmt.Errorf("error executing command: %v", err)
	}

	return handlerDiagnostics, nil
}

func TerraformApply(ctx context.Context, configMux *conflux.ConfigMux, preflight bool) (map[string]string, error) {
	terraformConfig := newTerraformConfig()
	diagnostics, err := conflux.Unmarshal(configMux, terraformConfig)
	if preflight {
		return diagnostics, nil
	} else if errors.Is(err, conflux.ErrInvalidFields) {
		return diagnostics, fmt.Errorf("error getting config for adding terraform config:\n%s", conflux.DiagnosticsToTable(diagnostics))
	} else if err != nil {
		return nil, fmt.Errorf("internal error unmarshalling terraform config to struct: %v", err)
	}

	terraformHandler := newTerraformHandler("./terraform/main.tf")
	handlerDiagnostics, err := terraformHandler.execute(ctx, terraformConfig)
	if err != nil {
		return nil, fmt.Errorf("error executing command: %v", err)
	}

	return handlerDiagnostics, nil
}
