package host

import (
	"fmt"
	"net/netip"
)

type SSHHost struct {
	Alias           string
	HostName        string
	User            string `json:"ssh_user" required:"true"`
	PublicKeyPath   string `json:"ssh_public_key_path" required:"true"`
	Port            string `json:"ssh_port" required:"true"`
	NodeCIDRAddress string `json:"node_cidr_address" required:"true"`
}

func NewSSHHost(hostAlias string) *SSHHost {
	return &SSHHost{
		Alias: hostAlias,
	}
}

func (c *SSHHost) FillInKeys() error {
	// ParsePrefix returns the prefix and an error if it's invalid
	prefix, err := netip.ParsePrefix(c.NodeCIDRAddress)
	if err != nil {
		return fmt.Errorf("'%s' is not a valid CIDR: %v", c.NodeCIDRAddress, err)
	}
	c.HostName = prefix.Addr().String()

	return nil
}
