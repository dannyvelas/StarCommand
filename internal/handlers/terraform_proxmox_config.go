package handlers

import "fmt"

type terraformProxmoxConfig struct {
	Node              string `json:"node" required:"true"`
	Endpoint          string `json:"endpoint" required:"true"`
	APIToken          string `json:"proxmox_terraform_user_api_token" required:"true"`
	SSHAddress        string `json:"ssh_address" required:"true"`
	SSHPort           string `json:"ssh_port" required:"true"`
	SSHPrivateKeyPath string `json:"ssh_private_key_path" required:"true"`
	TerraformVersion  string `json:"terraform_version" required:"true"`
	TerraformUsername string `json:"terraform_username" required:"true"`

	// injected
	UserRealmAPIToken string `json:"user_realm_api_token"`
}

func newTerraformProxmoxConfig() *terraformProxmoxConfig {
	return &terraformProxmoxConfig{
		SSHPort: "22",
	}
}

func (c *terraformProxmoxConfig) FillInKeys() error {
	c.UserRealmAPIToken = fmt.Sprintf("%s@pve!provider=%s", c.TerraformUsername, c.APIToken)
	return nil
}
