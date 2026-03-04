package app

import (
	"fmt"

	"github.com/dannyvelas/starcommand/config"
	"github.com/dannyvelas/starcommand/internal/helpers"
)

var _ ansiblePlaybookConfig = (*ansibleSetupVMConfig)(nil)

type setupVMEntry struct {
	Name string
	IP   string
	SSH  config.SSHConfig
}

type ansibleSetupVMConfig struct {
	VMs []setupVMEntry `json:"-" required:"true"`
}

func newAnsibleSetupVMConfig() *ansibleSetupVMConfig {
	return &ansibleSetupVMConfig{}
}

func (c *ansibleSetupVMConfig) FillFromConfig(cfg *config.Config) error {
	for _, host := range cfg.Hosts {
		for _, vm := range host.VMs {
			c.VMs = append(c.VMs, setupVMEntry{
				Name: vm.Name,
				IP:   vm.IP,
				SSH:  vm.SSH,
			})
		}
	}
	return nil
}

func (c *ansibleSetupVMConfig) generateHostVars() error {
	for _, vm := range c.VMs {
		if err := generateSetupVMVarsFor(vm); err != nil {
			return err
		}
	}
	return nil
}

func generateSetupVMVarsFor(entry setupVMEntry) error {
	ansibleUser, err := determineAnsibleUser(entry.SSH.User, entry.IP, portToString(entry.SSH.Port), entry.SSH.PrivateKeyPath)
	if err != nil {
		return fmt.Errorf("error determining ansible user for vm %s: %v", entry.Name, err)
	}

	expandedPrivateKey, err := helpers.ExpandPath(entry.SSH.PrivateKeyPath)
	if err != nil {
		return fmt.Errorf("error expanding private key path for vm %s: %v", entry.Name, err)
	}

	vars := setupVMHostVars{
		AnsibleHost:          entry.IP,
		AnsiblePort:          portToString(entry.SSH.Port),
		AnsibleSSHPrivateKey: expandedPrivateKey,
		AnsibleUser:          ansibleUser,
	}

	return writeHostVarsFile(entry.Name, vars)
}
