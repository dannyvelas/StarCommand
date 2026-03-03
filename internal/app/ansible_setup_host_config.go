package app

import "github.com/dannyvelas/starcommand/config"

var _ ansibleConfig = (*ansibleSetupHostConfig)(nil)

type ansibleSetupHostConfig struct {
	ansibleBaseConfig

	// Required
	IncusStoragePoolName string `json:"incus_storage_pool_name" required:"true"`
	IncusStorageDriver   string `json:"incus_storage_driver" required:"true"`

	// Sensitive
	SMTPUser     string `json:"smtp_user" sensitive:"true" prompt:"SMTP username"`
	SMTPPassword string `json:"smtp_password" sensitive:"true" prompt:"SMTP password"`
}

func newAnsibleSetupHostConfig() *ansibleSetupHostConfig {
	return &ansibleSetupHostConfig{
		ansibleBaseConfig: ansibleBaseConfig{
			SSHPort: "22",
		},
	}
}

func (c *ansibleSetupHostConfig) FillFromConfig(_ *config.Config) error { return nil }

func (c *ansibleSetupHostConfig) FillInKeys() error {
	return c.fillInBaseKeys()
}
