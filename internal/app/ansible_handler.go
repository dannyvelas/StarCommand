package app

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/dannyvelas/starcommand/internal/helpers"
)

type ansibleHandler struct{}

func newAnsibleHandler() ansibleHandler {
	return ansibleHandler{}
}

func (h ansibleHandler) getConfig(playbook string) (playbookConfig, error) {
	switch playbook {
	case "bootstrap-server":
		return newAnsibleBootstrapConfig(), nil
	case "setup-host":
		return newAnsibleSetupHostConfig(), nil
	case "setup-vm":
		return newAnsibleSetupVMConfig(), nil
	}

	return nil, fmt.Errorf("error: config for playbook %s %w", playbook, errNotFound)
}

func (h ansibleHandler) execute(cfg playbookConfig, playbook string) (map[string]string, error) {
	diagnostics := make(map[string]string)

	if err := cfg.generateHostVars(); err != nil {
		return diagnostics, fmt.Errorf("error generating host vars: %v", err)
	}

	if err := h.runAnsiblePlaybook(playbook); err != nil {
		return diagnostics, fmt.Errorf("error running ansible playbook: %v", err)
	}

	return diagnostics, nil
}

func (h ansibleHandler) runAnsiblePlaybook(playbook string) error {
	playbookPath := filepath.Join("ansible", "playbooks", playbook+".yml")
	args := []string{
		"-i", filepath.Join(".generated", "ansible", "inventory", "hosts.yml"),
		"-i", ".generated/",
		playbookPath,
	}

	cmd := exec.Command("ansible-playbook", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running ansible playbook: %v", err)
	}

	return nil
}

func determineAnsibleUser(sshUser, ip, port, privateKeyPath string) (string, error) {
	expandedKey, err := helpers.ExpandPath(privateKeyPath)
	if err != nil {
		return "", fmt.Errorf("error expanding key path: %v", err)
	}

	addr := fmt.Sprintf("%s:%s", ip, port)
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
