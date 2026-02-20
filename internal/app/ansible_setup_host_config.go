package app

import (
	"fmt"

	"github.com/dannyvelas/homelab/internal/helpers"
)

var _ ansibleConfig = (*ansibleSetupHostConfig)(nil)

type ansibleSetupHostConfig struct {
	// Required
	NodeIP               string `json:"node_ip" required:"true"`
	SSHUser              string `json:"ssh_user" required:"true"`
	SSHPort              string `json:"ssh_port" required:"true"`
	SSHPrivateKeyPath    string `json:"ssh_private_key_path" required:"true"`
	IncusStoragePoolName string `json:"incus_storage_pool_name" required:"true"`
	IncusStorageDriver   string `json:"incus_storage_driver" required:"true"`

	// Injected
	AnsibleUser string `json:"ansible_user"`
	AnsiblePort string `json:"ansible_port"`
}

// newAnsibleSetupHostConfig returns a pointer to an ansibleSetupHostConfig struct with some defaults
func newAnsibleSetupHostConfig() *ansibleSetupHostConfig {
	return &ansibleSetupHostConfig{
		SSHPort: "22",
	}
}

func (c *ansibleSetupHostConfig) FillInKeys() error {
	expandedPrivateKeyPath, err := helpers.ExpandPath(c.SSHPrivateKeyPath)
	if err != nil {
		return fmt.Errorf("error expanding path(%s): %v", c.SSHPrivateKeyPath, err)
	}
	c.SSHPrivateKeyPath = expandedPrivateKeyPath

	c.AnsibleUser = c.SSHUser
	c.AnsiblePort = c.SSHPort

	return nil
}

func (c *ansibleSetupHostConfig) GetNodeIP() string {
	return c.NodeIP
}

func (c *ansibleSetupHostConfig) GetSSHUser() string {
	return c.SSHUser
}

func (c *ansibleSetupHostConfig) GetSSHPort() string {
	return c.SSHPort
}

func (c *ansibleSetupHostConfig) GetSSHPrivateKeyPath() string {
	return c.SSHPrivateKeyPath
}
