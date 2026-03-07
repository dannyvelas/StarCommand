package app

import (
	"context"
	"fmt"
	"os"

	"github.com/dannyvelas/starcommand/internal/models"
)

func Setup(ctx context.Context, c *models.Config, hosts []string) error {
	return nil
}

func InventoryGenerate(ctx context.Context, c *models.Config, preflight bool) error {
	resolvedHostNames := resolveHostNames(c, nil)
	targets, err := resolveHosts(c, resolvedHostNames...)
	if err != nil {
		return fmt.Errorf("error resolving hosts: %v", err)
	}

	inventoryConfig := newInventoryConfig(targets)
	inventoryHandler := newInventoryHandler()
	if err := inventoryHandler.execute(inventoryConfig); err != nil {
		return fmt.Errorf("error executing command: %v", err)
	}

	return nil
}

func AnsibleRun(ctx context.Context, c *models.Config, playbook string, hosts []string, preflight bool) (*Diagnostics, error) {
	resolvedHostNames := resolveHostNames(c, hosts)
	targets, err := resolveHosts(c, resolvedHostNames...)
	if err != nil {
		return nil, fmt.Errorf("error resolving hosts: %v", err)
	}

	ansibleConfig, diagnostics, err := getAnsibleConfig(playbook, targets)
	if err != nil {
		return nil, fmt.Errorf("error getting config for %s: %v", playbook, err)
	}

	if preflight {
		return diagnostics, nil
	}

	if diagnostics.hasErrors() {
		return diagnostics, fmt.Errorf("config validation failed")
	}

	if err := promptSensitiveFields(ansibleConfig, os.Stdin, os.Stdout); err != nil {
		return nil, fmt.Errorf("error prompting for sensitive fields: %v", err)
	}

	ansibleHandler := newAnsibleHandler()
	if err := ansibleHandler.execute(ansibleConfig, playbook); err != nil {
		return nil, fmt.Errorf("error executing command: %v", err)
	}

	return nil, nil
}

func SSHAdd(ctx context.Context, c *models.Config, host string, preflight bool) error {
	targets, err := resolveHosts(c, host)
	if err != nil {
		return fmt.Errorf("error resolving hosts: %v", err)
	}

	sshConfig, err := newSSHConfig(targets[0])
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

func TerraformApply(ctx context.Context, c *models.Config, preflight bool) error {
	terraformConfig := newTerraformConfig()

	terraformHandler := newTerraformHandler("./terraform/main.tf")
	if err := terraformHandler.execute(ctx, terraformConfig); err != nil {
		return fmt.Errorf("error executing command: %v", err)
	}

	return nil
}

func resolveHosts(c *models.Config, hostNames ...string) ([]models.Host, error) {
	hostsNotFound := make([]string, 0)
	nameToHostMap := getNameToHostMap(c)

	configHosts := make([]models.Host, 0, len(hostNames))
	for _, hostName := range hostNames {
		host, found := nameToHostMap[hostName]
		if !found {
			hostsNotFound = append(hostsNotFound, hostName)
			continue
		}

		configHosts = append(configHosts, host)
	}

	if len(hostsNotFound) > 0 {
		return nil, fmt.Errorf("the following hosts were not found in the config: %v", hostsNotFound)
	}

	return configHosts, nil
}

func getNameToHostMap(c *models.Config) map[string]models.Host {
	nameToHostMap := make(map[string]models.Host, len(c.Hosts))
	for _, host := range c.Hosts {
		nameToHostMap[host.Name] = host
	}
	return nameToHostMap
}

// getHostList returns hosts if non-empty, otherwise returns the name of every
// non-VM host defined in c. This reflects the CLI semantics where omitting
// --host flags is equivalent to passing all hosts explicitly.
func resolveHostNames(c *models.Config, cliHosts []string) []string {
	if len(cliHosts) > 0 {
		return cliHosts
	}
	configHostNames := make([]string, 0, len(c.Hosts))
	for _, h := range c.Hosts {
		configHostNames = append(configHostNames, h.Name)
	}
	return configHostNames
}
