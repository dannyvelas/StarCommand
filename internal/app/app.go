package app

import (
	"context"
	"fmt"
	"os"

	"github.com/dannyvelas/starcommand/config"
)

func Setup(ctx context.Context, c *config.Config, hostAliases []string, preflight bool) (map[string]string, error) {
	return nil, nil
}

func InventoryGenerate(ctx context.Context, c *config.Config, host *string, preflight bool) (map[string]string, error) {
	return nil, nil
}

func AnsibleRun(ctx context.Context, c *config.Config, playbook string, preflight bool) (map[string]string, error) {
	ansibleHandler := newAnsibleHandler()

	playbookConfig := newAnsibleBootstrapConfig()

	m, err := loadConfig(playbookConfig, c)
	if err != nil {
		return m, err
	}

	if preflight {
		return m, nil
	}

	if hasMissingFields(m) {
		return m, fmt.Errorf("error getting config for %s:\n%s", playbook, diagnosticsToTable(m))
	}

	if err := promptSensitiveFields(playbookConfig, os.Stdin, os.Stdout); err != nil {
		return nil, fmt.Errorf("error prompting for sensitive fields: %v", err)
	}

	handlerDiagnostics, err := ansibleHandler.execute(playbookConfig, playbook)
	if err != nil {
		return nil, fmt.Errorf("error executing command: %v", err)
	}

	return handlerDiagnostics, nil
}

func SSHAdd(ctx context.Context, c *config.Config, hostAlias string, preflight bool) (map[string]string, error) {
	sshConfig := newSSHHost(hostAlias)
	//m, err := loadConfig(sshConfig, c, "ssh add "+hostAlias, preflight)
	//if preflight || err != nil {
	//	return m, err
	//}

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

func TerraformApply(ctx context.Context, c *config.Config, preflight bool) (map[string]string, error) {
	terraformConfig := newTerraformConfig()
	//m, err := loadConfig(terraformConfig, c, "terraform apply")
	//if preflight || err != nil {
	//	return m, err
	//}

	terraformHandler := newTerraformHandler("./terraform/main.tf")
	handlerDiagnostics, err := terraformHandler.execute(ctx, terraformConfig)
	if err != nil {
		return nil, fmt.Errorf("error executing command: %v", err)
	}

	return handlerDiagnostics, nil
}
