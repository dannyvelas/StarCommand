package app

import (
	"fmt"

	"github.com/dannyvelas/starcommand/internal/helpers"
)

// ansibleConfig is the full interface required to load and run a playbook.
type ansibleConfig interface {
	GetNodeIP() string
	GetSSHUser() string
	GetSSHPort() string
	GetSSHPrivateKeyPath() string
	configLoader
}

type ansibleBaseConfig struct {
	// Required
	NodeIP            string `json:"node_ip" required:"true"`
	SSHUser           string `json:"ssh_user" required:"true"`
	SSHPort           string `json:"ssh_port" required:"true"`
	SSHPrivateKeyPath string `json:"ssh_private_key_path" required:"true"`

	// Injected
	AnsibleUser string `json:"ansible_user"`
	AnsiblePort string `json:"ansible_port"`
}

func (c *ansibleBaseConfig) GetNodeIP() string            { return c.NodeIP }
func (c *ansibleBaseConfig) GetSSHUser() string           { return c.SSHUser }
func (c *ansibleBaseConfig) GetSSHPort() string           { return c.SSHPort }
func (c *ansibleBaseConfig) GetSSHPrivateKeyPath() string { return c.SSHPrivateKeyPath }

func (c *ansibleBaseConfig) fillInBaseKeys() error {
	expanded, err := helpers.ExpandPath(c.SSHPrivateKeyPath)
	if err != nil {
		return fmt.Errorf("error expanding path(%s): %v", c.SSHPrivateKeyPath, err)
	}
	c.SSHPrivateKeyPath = expanded
	c.AnsibleUser = c.SSHUser
	c.AnsiblePort = c.SSHPort
	return nil
}
