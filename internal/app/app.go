package app

import (
	"context"
	"fmt"
	"os"

	"github.com/dannyvelas/starcommand/internal/config"
)

func Setup(ctx context.Context, c *config.Config, hosts []string) error {
	return nil
}

func InventoryGenerate(ctx context.Context, c *config.Config) error {
	return nil
}

func AnsibleRun(ctx context.Context, c *config.Config, playbook string, hosts []string) error {
	playbookConfig, err := getAnsibleConfig(c, playbook)
	if err != nil {
		return fmt.Errorf("error getting config for %s: %v", playbook, err)
	}

	if err := promptSensitiveFields(playbookConfig, os.Stdin, os.Stdout); err != nil {
		return fmt.Errorf("error prompting for sensitive fields: %v", err)
	}

	ansibleHandler := newAnsibleHandler()
	if err := ansibleHandler.execute(playbookConfig, playbook); err != nil {
		return fmt.Errorf("error executing command: %v", err)
	}

	return nil
}

func SSHAdd(ctx context.Context, c *config.Config, host string) error {
	sshConfig, err := newSSHConfig(c, host)
	if err != nil {
		return fmt.Errorf("error creating ssh config: %v", err)
	}

	if err := promptSensitiveFields(sshConfig, os.Stdin, os.Stdout); err != nil {
		return fmt.Errorf("error prompting for sensitive fields: %v", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting user home dir: %v", err)
	}

	sshHandler := newSSHHandler(homeDir)
	if err := sshHandler.execute(sshConfig, host); err != nil {
		return fmt.Errorf("error executing command: %v", err)
	}

	return nil
}

func TerraformApply(ctx context.Context, c *config.Config) error {
	terraformConfig := newTerraformConfig()

	terraformHandler := newTerraformHandler("./terraform/main.tf")
	if err := terraformHandler.execute(ctx, terraformConfig); err != nil {
		return fmt.Errorf("error executing command: %v", err)
	}

	return nil
}
