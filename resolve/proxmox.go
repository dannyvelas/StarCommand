package resolve

import (
	"fmt"
	"net/netip"
	"os"
)

type proxmoxConfig struct {
	// Required
	SSHPublicKeyPath     string `yaml:"ssh_public_key_path" json:"ssh_public_key_path"`
	NodeCIDRAddress      string `yaml:"node_cidr_address" json:"node_cidr_address"`
	GatewayAddress       string `yaml:"gateway_address" json:"gateway_address"`
	PhysicalNIC          string `yaml:"physical_nic" json:"physical_nic"`
	AdminPassword        string `yaml:"admin_password" json:"admin_password"`
	SSHPort              string `yaml:"ssh_port" json:"ssh_port"`
	AutoUpdateRebootTime string `yaml:"auto_update_reboot_time" json:"auto_update_reboot_time"`
	AdminEmail           string `yaml:"admin_email" json:"admin_email"`
	SMTPUser             string `yaml:"smtp_user" json:"smtp_user"`
	SMTPPassword         string `yaml:"smtp_password" json:"smtp_password"`
	// Injected
	NodeIP       string `yaml:"node_ip" json:"node_ip"`
	SSHPublicKey string `yaml:"ssh_public_key" json:"ssh_public_key"`
}

// NewProxmoxConfig returns a pointer to the zero-value of ProxmoxConfig
func NewProxmoxConfig() *proxmoxConfig {
	return &proxmoxConfig{}
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

func (p *proxmoxConfig) FillInKeys() error {
	// ParsePrefix returns the prefix and an error if it's invalid
	if _, err := netip.ParsePrefix(p.NodeCIDRAddress); err != nil {
		return fmt.Errorf("'%s' is not a valid CIDR: %v", p.NodeCIDRAddress, err)
	}

	bytes, err := os.ReadFile(p.SSHPublicKeyPath)
	if err != nil {
		return fmt.Errorf("error reading ssh public key file: %v", err)
	}
	p.SSHPublicKey = string(bytes)

	return nil
}
