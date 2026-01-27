package models

import (
	"fmt"
	"net/netip"
)

type SSHHost struct {
	Alias           string `json:"alias"`
	HostName        string `json:"host_name"`
	User            string `json:"ssh_user" required:"true"`
	PublicKeyPath   string `json:"ssh_public_key_path" required:"true"`
	Port            string `json:"ssh_port" required:"true"`
	NodeCIDRAddress string `json:"node_cidr_address" required:"true"`
}

func NewSSHHost(hostAlias string) *SSHHost {
	return &SSHHost{Alias: hostAlias}
}

func (s *SSHHost) Name() string {
	return "ssh"
}

func (s *SSHHost) Resource() string {
	return "host"
}

func (s *SSHHost) FillInKeys() error {
	// ParsePrefix returns the prefix and an error if it's invalid
	prefix, err := netip.ParsePrefix(s.NodeCIDRAddress)
	if err != nil {
		return fmt.Errorf("'%s' is not a valid CIDR: %v", s.NodeCIDRAddress, err)
	}
	s.HostName = prefix.Addr().String()

	return nil
}
