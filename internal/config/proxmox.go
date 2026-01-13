package config

import (
	"fmt"
	"net/netip"
	"os"

	"github.com/dannyvelas/homelab/internal/helpers"
)

var _ config = (*proxmoxConfig)(nil)

type proxmoxConfig struct {
	// Required
	SSHPublicKeyPath     string `json:"ssh_public_key_path" required:"true"`
	NodeCIDRAddress      string `json:"node_cidr_address" required:"true"`
	GatewayAddress       string `json:"gateway_address" required:"true"`
	PhysicalNIC          string `json:"physical_nic" required:"true"`
	AdminPassword        string `json:"admin_password" required:"true" bw:"proxmox_admin_password"`
	SSHPort              string `json:"ssh_port" required:"true"`
	AutoUpdateRebootTime string `json:"auto_update_reboot_time" required:"true"`
	AdminEmail           string `json:"admin_email" required:"true"`
	SMTPUser             string `json:"smtp_user" required:"true"`
	SMTPPassword         string `json:"smtp_password" required:"true"`
	// Injected
	NodeIP       string `json:"node_ip"`
	SSHPublicKey string `json:"ssh_public_key"`
}

// NewProxmoxConfig returns a pointer to a ProxmoxConfig with some defaults
func newProxmoxConfig() *proxmoxConfig {
	return &proxmoxConfig{
		SSHPort:              "22",
		AutoUpdateRebootTime: "05:00",
	}
}

func (c *proxmoxConfig) Validate(diagnosticMap map[string]string) bool {
	ok := true
	if diagnosticMap["node_cidr_address"] == statusLoaded {
		if _, err := netip.ParsePrefix(c.NodeCIDRAddress); err != nil {
			diagnosticMap["node_cidr_address"] = fmt.Sprintf("'%s' is not a valid CIDR: %v\n", c.NodeCIDRAddress, err)
			ok = false
		}
	}

	return ok
}

func (c *proxmoxConfig) FillInKeys() error {
	// ParsePrefix returns the prefix and an error if it's invalid
	prefix, err := netip.ParsePrefix(c.NodeCIDRAddress)
	if err != nil {
		return fmt.Errorf("'%s' is not a valid CIDR: %v", c.NodeCIDRAddress, err)
	}
	c.NodeIP = prefix.Addr().String()

	expandedPublicKeyPath, err := helpers.ExpandPath(c.SSHPublicKeyPath)
	if err != nil {
		return fmt.Errorf("error expanding path(%s): %v", c.SSHPublicKeyPath, err)
	}

	bytes, err := os.ReadFile(expandedPublicKeyPath)
	if err != nil {
		return fmt.Errorf("error reading ssh public key file: %v", err)
	}
	c.SSHPublicKey = string(bytes)

	return nil
}
