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
	targets, err := resolveHosts(c, hosts)
	if err != nil {
		return fmt.Errorf("error resolving hosts: %v", err)
	}

	playbookConfig, err := getAnsibleConfig(playbook, targets)
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

func resolveHosts(c *config.Config, cliHosts []string) ([]config.Host, error) {
	hostsNotFound := make([]string, 0)
	hostNames := resolveHostNames(c, cliHosts)
	nameToHostMap := getNameToHostMap(c)

	configHosts := make([]config.Host, 0, len(hostNames))
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

func getNameToHostMap(c *config.Config) map[string]config.Host {
	nameToHostMap := make(map[string]config.Host, len(c.Hosts))
	for _, host := range c.Hosts {
		nameToHostMap[host.Name] = host
	}
	return nameToHostMap
}

// getHostList returns hosts if non-empty, otherwise returns the name of every
// non-VM host defined in c. This reflects the CLI semantics where omitting
// --host flags is equivalent to passing all hosts explicitly.
func resolveHostNames(c *config.Config, cliHosts []string) []string {
	if len(cliHosts) > 0 {
		return cliHosts
	}
	configHostNames := make([]string, 0, len(c.Hosts))
	for _, h := range c.Hosts {
		configHostNames = append(configHostNames, h.Name)
	}
	return configHostNames
}
