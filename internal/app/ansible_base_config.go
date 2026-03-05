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

	ansibleUser, err := determineAnsibleUser(sshUser, ip, sshPort, expandedPrivateKey)
	if err != nil {
		return ansibleHostConfig{}, fmt.Errorf("error determining ansible user for %s: %v", name, err)
	}

	return ansibleHostConfig{
		Name: name,
		Map: map[string]any{
			"ansible_host": ip,
			"ansible_port": sshPort,
			"ansible_user": ansibleUser,
		},
	}, nil
}
