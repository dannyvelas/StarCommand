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

func newAnsibleBaseConfig(name, ip, sshUser string, sshPort int, sshPrivateKeyPath string) ansibleBaseConfig {
	return ansibleBaseConfig{
		Name:              name,
		IP:                ip,
		SSHUser:           sshUser,
		SSHPort:           sshPort,
		SSHPrivateKeyPath: sshPrivateKeyPath,
	}
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
