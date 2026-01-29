package handlers

import (
	"fmt"
	"net/netip"
)

type sshConfig struct {
	Alias           string `json:"alias"`
	HostName        string `json:"host_name"`
	User            string `json:"ssh_user" required:"true"`
	PublicKeyPath   string `json:"ssh_public_key_path" required:"true"`
	Port            string `json:"ssh_port" required:"true"`
	NodeCIDRAddress string `json:"node_cidr_address" required:"true"`
}

func newSSHHost(hostAlias string) *sshConfig {
	return &sshConfig{Alias: hostAlias}
}

func (s *sshConfig) FillInKeys() error {
	// ParsePrefix returns the prefix and an error if it's invalid
	prefix, err := netip.ParsePrefix(s.NodeCIDRAddress)
	if err != nil {
		return fmt.Errorf("'%s' is not a valid CIDR: %v", s.NodeCIDRAddress, err)
	}
	s.HostName = prefix.Addr().String()

	return nil
}
