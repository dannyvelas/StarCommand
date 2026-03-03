package app

import "github.com/dannyvelas/starcommand/config"

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

func (c *ansibleSetupVMConfig) FillFromConfig(_ *config.Config) error { return nil }

func (c *ansibleSetupVMConfig) FillInKeys() error {
	return c.fillInBaseKeys()
}
