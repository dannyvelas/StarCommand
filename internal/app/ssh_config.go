package app

import (
	"fmt"
	"slices"

	"github.com/dannyvelas/starcommand/config"
)

type sshConfig struct {
	Alias         string `json:"alias"`
	HostName      string `json:"host_name"`
	User          string `json:"ssh_user"`
	PublicKeyPath string `json:"ssh_public_key_path"`
	Port          int    `json:"ssh_port"`
}

func newSSHConfig(c *config.Config, hostAlias string) (*sshConfig, map[string]string, error) {
	i := slices.IndexFunc(c.Hosts, func(h config.Host) bool { return h.Name == hostAlias })
	if i < 0 {
		return nil, nil, fmt.Errorf("host alias %s not found in config", hostAlias)
	}
	configHost := c.Hosts[i]

	diagnostics := make(map[string]string)
	prefix := fmt.Sprintf(".hosts[%d]", i)
	setDiagnostic(diagnostics, prefix+".name", configHost.Name)
	setDiagnostic(diagnostics, prefix+".ip", configHost.IP)
	setDiagnostic(diagnostics, prefix+".ssh.user", configHost.SSH.User)
	setDiagnostic(diagnostics, prefix+".ssh.public_key_path", configHost.SSH.PublicKeyPath)
	setDiagnostic(diagnostics, prefix+".ssh.port", configHost.SSH.Port)

	return &sshConfig{
		Alias:         configHost.Name,
		HostName:      configHost.IP,
		User:          configHost.SSH.User,
		PublicKeyPath: configHost.SSH.PublicKeyPath,
		Port:          configHost.SSH.Port,
	}, diagnostics, nil
}
