package config

import (
	"fmt"
	"net/netip"
	"os"

	"github.com/dannyvelas/homelab/internal/helpers"
)

type proxmoxConfig struct {
	// Required
	SSHPublicKeyPath     string `json:"ssh_public_key_path" required:"true"`
	NodeCIDRAddress      string `json:"node_cidr_address" required:"true"`
	GatewayAddress       string `json:"gateway_address" required:"true"`
	PhysicalNIC          string `json:"physical_nic" required:"true"`
	AdminPassword        string `json:"admin_password" required:"true" bw:"proxmox_admin_password" `
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

func (p *proxmoxConfig) Validate(results map[string]string) bool {
	ok := true
	if results["node_cidr_address"] == statusLoaded {
		if _, err := netip.ParsePrefix(p.NodeCIDRAddress); err != nil {
			results["node_cidr_address"] = fmt.Sprintf("'%s' is not a valid CIDR: %v\n", p.NodeCIDRAddress, err)
			ok = false
		}
	}

	return ok
}

func (p *proxmoxConfig) FillInKeys() error {
	// ParsePrefix returns the prefix and an error if it's invalid
	prefix, err := netip.ParsePrefix(p.NodeCIDRAddress)
	if err != nil {
		return fmt.Errorf("'%s' is not a valid CIDR: %v", p.NodeCIDRAddress, err)
	}
	p.NodeIP = prefix.Addr().String()

	expandedPublicKeyPath, err := helpers.ExpandPath(p.SSHPublicKeyPath)
	if err != nil {
		return fmt.Errorf("error expanding path(%s): %v", p.SSHPublicKeyPath, err)
	}

	bytes, err := os.ReadFile(expandedPublicKeyPath)
	if err != nil {
		return fmt.Errorf("error reading ssh public key file: %v", err)
	}
	p.SSHPublicKey = string(bytes)

	return nil
}
