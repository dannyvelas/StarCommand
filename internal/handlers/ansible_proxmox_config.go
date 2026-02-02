package handlers

import (
	"fmt"
	"net/netip"
	"os"

	"github.com/dannyvelas/conflux"
	"github.com/dannyvelas/homelab/internal/helpers"
)

type ansibleProxmoxConfig struct {
	// Required
	SSHPrivateKeyPath    string `json:"ssh_private_key_path" required:"true"`
	SSHPublicKeyPath     string `json:"ssh_public_key_path" required:"true"`
	NodeCIDRAddress      string `json:"node_cidr_address" required:"true"`
	GatewayAddress       string `json:"gateway_address" required:"true"`
	PhysicalNIC          string `json:"physical_nic" required:"true"`
	SSHUser              string `json:"ssh_user" required:"true"`
	SSHPort              string `json:"ssh_port" required:"true"`
	AutoUpdateRebootTime string `json:"auto_update_reboot_time" required:"true"`
	AdminEmail           string `json:"admin_email" required:"true"`
	AdminPassword        string `json:"admin_password" required:"true" conflux:"proxmox_admin_password"`
	SMTPUser             string `json:"smtp_user" required:"true"`
	SMTPPassword         string `json:"smtp_password" required:"true"`
	TerraformUsername    string `json:"terraform_username" required:"true"`

	// fields needed to upload secrets to bitwarden
	BitwardenAPIURL         string `json:"bitwarden_api_url" required:"true"`
	BitwardenIdentityURL    string `json:"bitwarden_identity_url" required:"true"`
	BitwardenAccessToken    string `json:"bitwarden_access_token" required:"true"`
	BitwardenProjectID      string `json:"bitwarden_project_id" required:"true"`
	BitwardenOrganizationID string `json:"bitwarden_organization_id" required:"true"`
	BitwardenStateFilePath  string `json:"bitwarden_state_file_path" required:"true"`

	// Injected
	NodeIP       string `json:"node_ip"`
	SSHPublicKey string `json:"ssh_public_key"`
	AnsibleUser  string `json:"ansible_user"`
	AnsiblePort  string `json:"ansible_port"`
}

// NewAnsibleProxmoxConfig returns a pointer to a Proxmox struct with some defaults
func newAnsibleProxmoxConfig() *ansibleProxmoxConfig {
	return &ansibleProxmoxConfig{
		SSHPort:                "22",
		AutoUpdateRebootTime:   "05:00",
		BitwardenAPIURL:        "https://api.bitwarden.com",
		BitwardenIdentityURL:   "https://identity.bitwarden.com",
		BitwardenStateFilePath: ".bw_state",
	}
}

func (c *ansibleProxmoxConfig) Validate(diagnostics map[string]string) bool {
	ok := true
	if diagnostics["node_cidr_address"] == conflux.StatusLoaded {
		if _, err := netip.ParsePrefix(c.NodeCIDRAddress); err != nil {
			diagnostics["node_cidr_address"] = fmt.Sprintf("'%s' is not a valid CIDR: %v\n", c.NodeCIDRAddress, err)
			ok = false
		}
	}

	return ok
}

func (c *ansibleProxmoxConfig) FillInKeys() error {
	// ParsePrefix returns the prefix and an error if it's invalid
	prefix, err := netip.ParsePrefix(c.NodeCIDRAddress)
	if err != nil {
		return fmt.Errorf("'%s' is not a valid CIDR: %v", c.NodeCIDRAddress, err)
	}
	c.NodeIP = prefix.Addr().String()

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
