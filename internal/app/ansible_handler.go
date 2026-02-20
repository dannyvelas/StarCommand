package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"golang.org/x/crypto/ssh"
)

type ansibleHandler struct{}

func newAnsibleHandler() ansibleHandler {
	return ansibleHandler{}
}

func (h ansibleHandler) getConfig(playbook string) (ansibleConfig, error) {
	switch playbook {
	case "bootstrap-server":
		return newAnsibleBootstrapConfig(), nil
	case "setup-host":
		return newAnsibleSetupHostConfig(), nil
	case "setup-vm":
	}

	return nil, fmt.Errorf("error: config for playbook %s %w", playbook, errNotFound)
}

func (h ansibleHandler) execute(config ansibleConfig) (map[string]string, error) {
	diagnostics := make(map[string]string)

	if err := h.runAnsiblePlaybook(config); err != nil {
		return diagnostics, fmt.Errorf("error running ansible playbook: %v", err)
	}

	return diagnostics, nil
}

func (h ansibleHandler) runAnsiblePlaybook(config ansibleConfig) error {
	proxmoxAddr := fmt.Sprintf("%s:%s", config.GetNodeIP(), config.GetSSHPort())
	client, sshErr := h.getSSHClient(config.GetSSHUser(), proxmoxAddr, config.GetSSHPrivateKeyPath())
	if sshErr != nil && !errors.Is(sshErr, errConnectingSSH) {
		return fmt.Errorf("error checking if ssh is accessible to proxmox host: %v", sshErr)
	} else if sshErr == nil {
		_ = client.Close()
	}

	tmpFile, err := os.CreateTemp("", "labctl-vars-*.json")
	if err != nil {
		return fmt.Errorf("error creating temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	if err := json.NewEncoder(tmpFile).Encode(config); err != nil {
		return fmt.Errorf("error writing config to tmp file: %v", err)
	}

	args := []string{"-i", "ansible/inventory.ini", "ansible/setup-proxmox.yml", "-e", "@" + tmpFile.Name()}
	if errors.Is(sshErr, errConnectingSSH) {
		args = append(args, "-u", "root")
	}

	cmd := exec.Command("ansible-playbook", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running ansible proxmox command: %v", err)
	}

	return nil
}

func (h ansibleHandler) getSSHClient(user, addr, privateKeyPath string) (*ssh.Client, error) {
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
		Timeout:         3 * time.Second,             // -o ConnectTimeout=3
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // StrictHostKeyChecking=no
	}

	client, err := ssh.Dial("tcp", addr, sshClientConfig)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errConnectingSSH, err)
	}

	return client, nil
}
