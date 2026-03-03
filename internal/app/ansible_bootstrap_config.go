package app

import (
	"fmt"
	"os"

	"github.com/dannyvelas/starcommand/config"
	"github.com/dannyvelas/starcommand/internal/helpers"
)

var _ ansibleConfig = (*ansibleBootstrapConfig)(nil)

type ansibleBootstrapConfig struct {
	ansibleBaseConfig

	// Required
	SSHPublicKeyPath     string `json:"ssh_public_key_path" required:"true"`
	AutoUpdateRebootTime string `json:"auto_update_reboot_time" required:"true"`

	// Injected
	SSHPublicKey string `json:"ssh_public_key"`

	// Sensitive
	AdminEmail    string `json:"admin_email" sensitive:"true" prompt:"Admin email"`
	AdminPassword string `json:"admin_password" sensitive:"true" prompt:"Admin password"`
}

func newAnsibleBootstrapConfig() *ansibleBootstrapConfig {
	return &ansibleBootstrapConfig{
		ansibleBaseConfig: ansibleBaseConfig{
			SSHPort: "22",
		},
		AutoUpdateRebootTime: "05:00",
	}
}

func (c *ansibleBootstrapConfig) FillFromConfig(_ *config.Config) error { return nil }

func (c *ansibleBootstrapConfig) FillInKeys() error {
	if err := c.fillInBaseKeys(); err != nil {
		return err
	}

	expandedPublicKeyPath, err := helpers.ExpandPath(c.SSHPublicKeyPath)
	if err != nil {
		return fmt.Errorf("error expanding path(%s): %v", c.SSHPublicKeyPath, err)
	}
	c.SSHPublicKeyPath = expandedPublicKeyPath

	bytes, err := os.ReadFile(expandedPublicKeyPath)
	if err != nil {
		return fmt.Errorf("error reading ssh public key file: %v", err)
	}
	c.SSHPublicKey = string(bytes)

	return nil
}
