package app

import (
	"fmt"

	"github.com/dannyvelas/starcommand/config"
)

var _ ansibleConfig = (*ansibleSetupVMConfig)(nil)

type ansibleSetupVMConfig struct {
	ansibleBaseConfig
}

func newAnsibleSetupVMConfig() *ansibleSetupVMConfig {
	return &ansibleSetupVMConfig{
		ansibleBaseConfig: ansibleBaseConfig{
			SSHPort: "22",
		},
	}
}

func (c *ansibleSetupVMConfig) FillFromConfig(cfg *config.Config) error {
	if len(cfg.Hosts) == 0 {
		return fmt.Errorf("no hosts configured")
	}
	if len(cfg.Hosts[0].VMs) == 0 {
		return fmt.Errorf("no VMs configured for host %q", cfg.Hosts[0].Name)
	}
	c.fillBaseFromVM(cfg.Hosts[0].VMs[0])
	return nil
}

func (c *ansibleSetupVMConfig) FillInKeys() error {
	return c.fillInBaseKeys()
}
