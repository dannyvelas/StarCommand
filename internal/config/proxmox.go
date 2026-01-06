package config

import (
	"fmt"
	"net/netip"
	"os"

	"github.com/dannyvelas/homelab/internal/helpers"
)

type proxmoxConfig struct {
	// Required
	SSHPublicKeyPath     string `json:"ssh_public_key_path"`
	NodeCIDRAddress      string `json:"node_cidr_address"`
	GatewayAddress       string `json:"gateway_address"`
	PhysicalNIC          string `json:"physical_nic"`
	AdminPassword        string `json:"admin_password" bw:"proxmox_admin_password"`
	SSHPort              string `json:"ssh_port"`
	AutoUpdateRebootTime string `json:"auto_update_reboot_time"`
	AdminEmail           string `json:"admin_email"`
	SMTPUser             string `json:"smtp_user"`
	SMTPPassword         string `json:"smtp_password"`
	// Injected
	NodeIP       string `json:"node_ip"`
	SSHPublicKey string `json:"ssh_public_key"`
}

// NewProxmoxConfig returns a pointer to a ProxmoxConfig with some defaults
func NewProxmoxConfig() *proxmoxConfig {
	return &proxmoxConfig{
		SSHPort:              "22",
		AutoUpdateRebootTime: "05:00",
	}
}

func (p *proxmoxConfig) Validate() map[string]string {
	keyErrors := make(map[string]string)
	if p.SSHPublicKeyPath == "" {
		keyErrors["ssh_public_key_path"] = errMissing
	}

	if p.NodeCIDRAddress == "" {
		keyErrors["node_cidr_address"] = errMissing
	} else if _, err := netip.ParsePrefix(p.NodeCIDRAddress); err != nil {
		keyErrors["node_cidr_address"] = fmt.Sprintf("'%s' is not a valid CIDR: %v\n", p.NodeCIDRAddress, err)
	}

	if p.GatewayAddress == "" {
		keyErrors["gateway_address"] = errMissing
	}

	if p.PhysicalNIC == "" {
		keyErrors["physical_nic"] = errMissing
	}

	if p.AdminPassword == "" {
		keyErrors["admin_password"] = errMissing
	}

	if p.SSHPort == "" {
		keyErrors["ssh_port"] = errMissing
	}

	if p.AutoUpdateRebootTime == "" {
		keyErrors["auto_update_reboot_time"] = errMissing
	}

	if p.AdminEmail == "" {
		keyErrors["admin_email"] = errMissing
	}

	if p.SMTPUser == "" {
		keyErrors["smtp_user"] = errMissing
	}

	if p.SMTPPassword == "" {
		keyErrors["smtp_password"] = errMissing
	}

	return keyErrors
}

func (p *proxmoxConfig) RequiredKeys() []string {
	return []string{
		"ssh_public_key_path",
		"node_cidr_address",
		"gateway_address",
		"physical_nic",
		"proxmox_admin_password",
		"ssh_port",
		"auto_update_reboot_time",
		"admin_email",
		"smtp_user",
		"smtp_password",
	}
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
