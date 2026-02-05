package handlers

import (
	"fmt"

	"github.com/dannyvelas/homelab/internal/helpers"
)

type terraformPlexConfig struct {
	Endpoint                   string `json:"endpoint" required:"true"`
	GatewayAddress             string `json:"gateway_address" required:"true"`
	IP                         string `json:"ip" required:"true"`
	Node                       string `json:"node" required:"true"`
	SSHUser                    string `json:"ssh_user" required:"true"`
	SSHRealm                   string `json:"ssh_realm" required:"true"`
	Password                   string `json:"password" required:"true" conflux:"proxmox_root_password"`
	SSHPublicKeyPath           string `json:"ssh_public_key_path" required:"true"`
	TerraformVersionConstraint string `json:"-" required:"true" conflux:"terraform_version_constraint"`

	// injected
	UserRealm string `json:"user_realm"`
}

func newTerraformPlexConfig() *terraformPlexConfig {
	return &terraformPlexConfig{
		SSHUser:  "root",
		SSHRealm: "pam",
	}
}

func (c *terraformPlexConfig) FillInKeys() error {
	expandedPath, err := helpers.ExpandPath(c.SSHPublicKeyPath)
	if err != nil {
		return err
	}
	c.SSHPublicKeyPath = expandedPath

	c.UserRealm = fmt.Sprintf("%s@%s", c.SSHUser, c.SSHRealm)
	return nil
}
