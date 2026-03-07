package app

import (
	"fmt"

	"github.com/dannyvelas/starcommand/internal/helpers"
)

func newAnsibleBaseConfig(name, ip, sshUser string, sshPort int, sshPrivateKeyPath string) (ansibleHostConfig, error) {
	expandedPrivateKey, err := helpers.ExpandPath(sshPrivateKeyPath)
	if err != nil {
		return ansibleHostConfig{}, fmt.Errorf("error expanding private key path for %s: %v", name, err)
	}

	return ansibleHostConfig{
		Name:          name,
		IP:            ip,
		SSHUser:       sshUser,
		SSHPort:       sshPort,
		SSHPrivateKey: expandedPrivateKey,
	}, nil
}
