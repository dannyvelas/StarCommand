package app

import (
	"fmt"
	"os"

	"github.com/dannyvelas/homelab/internal/helpers"
)

var _ ansibleConfig = (*ansibleBootstrapConfig)(nil)

type ansibleBootstrapConfig struct {
	// Required
	NodeIP               string `json:"node_ip" required:"true"`
	SSHUser              string `json:"ssh_user" required:"true"`
	SSHPort              string `json:"ssh_port" required:"true"`
	SSHPrivateKeyPath    string `json:"ssh_private_key_path" required:"true"`
	SSHPublicKeyPath     string `json:"ssh_public_key_path" required:"true"`
	AutoUpdateRebootTime string `json:"auto_update_reboot_time" required:"true"`
	AdminEmail           string `json:"admin_email" required:"true"`
	AdminPassword        string `json:"admin_password" required:"true"`

	// Injected
	SSHPublicKey string `json:"ssh_public_key"`
	AnsibleUser  string `json:"ansible_user"`
	AnsiblePort  string `json:"ansible_port"`
}

// newAnsibleBootstrapConfig returns a pointer to an ansibleBootstrapConfig struct with some defaults
func newAnsibleBootstrapConfig() *ansibleBootstrapConfig {
	return &ansibleBootstrapConfig{
		SSHPort:              "22",
		AutoUpdateRebootTime: "05:00",
	}
}

func (c *ansibleBootstrapConfig) FillInKeys() error {
	expandedPrivateKeyPath, err := helpers.ExpandPath(c.SSHPrivateKeyPath)
	if err != nil {
		return fmt.Errorf("error expanding path(%s): %v", c.SSHPrivateKeyPath, err)
	}
	c.SSHPrivateKeyPath = expandedPrivateKeyPath

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

	c.AnsibleUser = c.SSHUser
	c.AnsiblePort = c.SSHPort

	return nil
}

func (c *ansibleBootstrapConfig) GetNodeIP() string {
	return c.NodeIP
}

func (c *ansibleBootstrapConfig) GetSSHUser() string {
	return c.SSHUser
}

func (c *ansibleBootstrapConfig) GetSSHPort() string {
	return c.SSHPort
}

func (c *ansibleBootstrapConfig) GetSSHPrivateKeyPath() string {
	return c.SSHPrivateKeyPath
}
