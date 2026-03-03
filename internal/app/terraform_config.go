package app

import (
	"fmt"

	"github.com/dannyvelas/starcommand/config"
	"github.com/dannyvelas/starcommand/internal/helpers"
)

type terraformConfig struct {
	Endpoint                   string `json:"endpoint" required:"true"`
	GatewayAddress             string `json:"gateway_address" required:"true"`
	IP                         string `json:"ip" required:"true"`
	Node                       string `json:"node" required:"true"`
	SSHUser                    string `json:"-" required:"true" conflux:"ssh_user"`
	SSHRealm                   string `json:"-" required:"true" conflux:"ssh_realm"`
	Password                   string `json:"password" required:"true" conflux:"proxmox_root_password"`
	SSHPublicKeyPath           string `json:"ssh_public_key_path" required:"true"`
	TerraformVersionConstraint string `json:"-" required:"true" conflux:"terraform_version_constraint"`

	// injected
	UserRealm string `json:"user_realm"`
}

func newTerraformConfig() *terraformConfig {
	return &terraformConfig{
		SSHUser:  "root",
		SSHRealm: "pam",
	}
}

func (c *terraformConfig) FillFromConfig(_ *config.Config) error { return nil }

func (c *terraformConfig) FillInKeys() error {
	expandedPath, err := helpers.ExpandPath(c.SSHPublicKeyPath)
	if err != nil {
		return err
	}
	c.SSHPublicKeyPath = expandedPath

	c.UserRealm = fmt.Sprintf("%s@%s", c.SSHUser, c.SSHRealm)
	return nil
}
