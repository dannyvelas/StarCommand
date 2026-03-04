package app

import (
	"fmt"

	"github.com/dannyvelas/starcommand/internal/helpers"
)

type ansibleBaseConfig struct {
	Name              string
	IP                string
	SSHUser           string
	SSHPort           int
	SSHPrivateKeyPath string
}

func newAnsibleBaseConfig(prefix, name, ip, sshUser string, sshPort int, sshPrivateKeyPath string) (ansibleBaseConfig, map[string]string) {
	diagnostics := make(map[string]string)
	setDiagnostic(diagnostics, prefix+".name", name)
	setDiagnostic(diagnostics, prefix+".ip", ip)
	setDiagnostic(diagnostics, prefix+".ssh.user", sshUser)
	setDiagnostic(diagnostics, prefix+".ssh.port", sshPort)
	setDiagnostic(diagnostics, prefix+".ssh.private_key_path", sshPrivateKeyPath)

	return ansibleBaseConfig{
			Name:              name,
			IP:                ip,
			SSHUser:           sshUser,
			SSHPort:           sshPort,
			SSHPrivateKeyPath: sshPrivateKeyPath,
		},
		diagnostics
}

func (c ansibleBaseConfig) asMap() (map[string]any, error) {
	expandedPrivateKey, err := helpers.ExpandPath(c.SSHPrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("error expanding private key path for %s: %v", c.Name, err)
	}

	ansibleUser, err := determineAnsibleUser(c.SSHUser, c.IP, c.SSHPort, expandedPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("error determining ansible user for %s: %v", c.Name, err)
	}

	return map[string]any{
		"ansible_host": c.IP,
		"ansible_port": c.SSHPort,
		"ansible_user": ansibleUser,
	}, nil
}
