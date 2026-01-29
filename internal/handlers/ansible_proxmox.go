package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"golang.org/x/crypto/ssh"
)

var errConnectingSSH = errors.New("error connecting via ssh")

var _ Handler = AnsibleProxmoxHandler{}

type AnsibleProxmoxHandler struct{}

func NewAnsibleProxmoxHandler() AnsibleProxmoxHandler {
	return AnsibleProxmoxHandler{}
}

func (h AnsibleProxmoxHandler) GetConfig(_ string) any {
	return newAnsibleProxmoxConfig()
}

func (h AnsibleProxmoxHandler) Execute(config any, hostAlias string) (map[string]string, error) {
	ansibleProxmoxConfig, ok := config.(*ansibleProxmoxConfig)
	if !ok {
		return nil, fmt.Errorf("internal type error converting config to ansible proxmox config. found: %T", config)
	}

	execCommand, err := h.getCommand(ansibleProxmoxConfig)
	if err != nil {
		return nil, fmt.Errorf("error determining if to use root for ansible playbook: %v", err)
	}

	if err := execCommand.Run(); err != nil {
		return nil, fmt.Errorf("error running ansible proxmox command: %v", err)
	}

	return nil, nil
}

// getCommand returns the command to run ansible, inferring whether to use root permissions
func (h AnsibleProxmoxHandler) getCommand(config *ansibleProxmoxConfig) (*exec.Cmd, error) {
	err := h.checkSSH(config)
	if err != nil && !errors.Is(err, errConnectingSSH) {
		return nil, fmt.Errorf("error checking if ssh is accessible to proxmox host: %v", err)
	}

	asJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("error converting config to JSON: %v", err)
	}

	if errors.Is(err, errConnectingSSH) {
		return exec.Command("ansible-playbook", "-i", "ansible/inventory.ini", "ansible/setup-proxmox.yml", "-u", "root", "-e", string(asJSON)), nil
	} else {
		return exec.Command("ansible-playbook", "-i", "ansible/inventory.ini", "ansible/setup-proxmox.yml", "-e", string(asJSON)), nil
	}
}

func (h AnsibleProxmoxHandler) checkSSH(config *ansibleProxmoxConfig) error {
	key, err := os.ReadFile(config.SSHPrivateKeyPath)
	if err != nil {
		return fmt.Errorf("unable to read private key: %v", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return fmt.Errorf("unable to parse private key: %v", err)
	}

	sshClientConfig := &ssh.ClientConfig{
		User:            config.SSHUser,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		Timeout:         3 * time.Second,             // -o ConnectTimeout=3
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // StrictHostKeyChecking=no
	}

	addr := fmt.Sprintf("%s:%s", config.NodeIP, config.SSHPort)
	client, err := ssh.Dial("tcp", addr, sshClientConfig)
	if err != nil {
		return fmt.Errorf("%w: %v", errConnectingSSH, err)
	}
	defer client.Close()

	return nil
}
