package app

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/dannyvelas/starcommand/internal/helpers"
	"golang.org/x/crypto/ssh"
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

func determineAnsibleUser(sshUser, ip string, port int, privateKeyPath string) (string, error) {
	expandedKey, err := helpers.ExpandPath(privateKeyPath)
	if err != nil {
		return "", fmt.Errorf("error expanding key path: %v", err)
	}

	addr := fmt.Sprintf("%s:%d", ip, port)
	client, sshErr := getSSHClient(sshUser, addr, expandedKey)
	if sshErr != nil && !errors.Is(sshErr, errConnectingSSH) {
		return "", fmt.Errorf("error checking ssh: %v", sshErr)
	} else if sshErr == nil {
		_ = client.Close()
		return sshUser, nil
	}

	return "root", nil
}

func getSSHClient(user, addr, privateKeyPath string) (*ssh.Client, error) {
	key, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read private key: %v", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("unable to parse private key: %v", err)
	}

	sshClientConfig := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		Timeout:         3 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", addr, sshClientConfig)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errConnectingSSH, err)
	}

	return client, nil
}
