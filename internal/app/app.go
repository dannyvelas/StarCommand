package app

import (
	"context"
	"fmt"
	"os"

	"github.com/dannyvelas/starcommand/config"
)

func Setup(ctx context.Context, c *config.Config, hostAliases []string) (map[string]string, error) {
	return nil, nil
}

func InventoryGenerate(ctx context.Context, c *config.Config, host *string) (map[string]string, error) {
	return nil, nil
}

func AnsibleRun(ctx context.Context, c *config.Config, playbook string) (map[string]string, error) {
	playbookConfig, err := getAnsibleConfig(c, playbook)
	if err != nil {
		return nil, fmt.Errorf("error getting config for %s: %v", playbook, err)
	}

	if err := promptSensitiveFields(playbookConfig, os.Stdin, os.Stdout); err != nil {
		return nil, fmt.Errorf("error prompting for sensitive fields: %v", err)
	}

	ansibleHandler := newAnsibleHandler()
	handlerDiagnostics, err := ansibleHandler.execute(playbookConfig, playbook)
	if err != nil {
		return nil, fmt.Errorf("error executing command: %v", err)
	}

	return handlerDiagnostics, nil
}

func SSHAdd(ctx context.Context, c *config.Config, hostAlias string) (map[string]string, error) {
	sshConfig, err := newSSHConfig(c, hostAlias)
	if err != nil {
		return nil, fmt.Errorf("error creating ssh config: %v", err)
	}

	if err := promptSensitiveFields(sshConfig, os.Stdin, os.Stdout); err != nil {
		return nil, fmt.Errorf("error prompting for sensitive fields: %v", err)
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

func TerraformApply(ctx context.Context, c *config.Config) (map[string]string, error) {
	terraformConfig := newTerraformConfig()

	terraformHandler := newTerraformHandler("./terraform/main.tf")
	handlerDiagnostics, err := terraformHandler.execute(ctx, terraformConfig)
	if err != nil {
		return nil, fmt.Errorf("error executing command: %v", err)
	}

	return handlerDiagnostics, nil
}
