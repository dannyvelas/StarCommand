package app

import (
	"fmt"

	"github.com/dannyvelas/starcommand/config"
	"github.com/dannyvelas/starcommand/internal/helpers"
)

type ansibleBaseConfig struct {
	Name string
	IP   string
	SSH  config.SSHConfig
}

func newAnsibleBaseConfig(name, ip string, ssh config.SSHConfig) ansibleBaseConfig {
	return ansibleBaseConfig{
		Name: name,
		IP:   ip,
		SSH:  ssh,
	}
}

func (c ansibleBaseConfig) asMap() (map[string]any, error) {
	expandedPrivateKey, err := helpers.ExpandPath(c.SSH.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("error expanding private key path for %s: %v", c.Name, err)
	}

	ansibleUser, err := determineAnsibleUser(c.SSH.User, c.IP, c.SSH.Port, expandedPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("error determining ansible user for %s: %v", c.Name, err)
	}

	return map[string]any{
		"ansible_host": c.IP,
		"ansible_port": c.SSH.Port,
		"ansible_user": ansibleUser,
	}, nil
}
