package app

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/dannyvelas/starcommand/internal/helpers"
	"github.com/goccy/go-yaml"
	"golang.org/x/crypto/ssh"
)

type ansibleHandler struct{}

func newAnsibleHandler() ansibleHandler {
	return ansibleHandler{}
}

func (h ansibleHandler) execute(c playbookConfig, playbook string) error {
	if err := h.generateHostVars(c); err != nil {
		return fmt.Errorf("error generating host vars: %v", err)
	}

	if err := h.runAnsiblePlaybook(playbook); err != nil {
		return fmt.Errorf("error running ansible playbook: %v", err)
	}

	return nil
}

func (h ansibleHandler) generateHostVars(c playbookConfig) error {
	for _, host := range c.hosts() {
		ansibleUser, err := determineAnsibleUser(host.SSHUser, host.IP, host.SSHPort, host.SSHPrivateKey)
		if err != nil {
			return fmt.Errorf("error determining ansible user for %s: %v", host.Name, err)
		}

		vars := make(map[string]any, len(host.Map)+3)
		maps.Copy(vars, host.Map)
		vars["ansible_host"] = host.IP
		vars["ansible_port"] = host.SSHPort
		vars["ansible_user"] = ansibleUser

		if err := h.writeHostVarsFile(host.Name, vars); err != nil {
			return fmt.Errorf("error writing host vars file for %s: %v", host.Name, err)
		}
	}

	return nil
}

func (h ansibleHandler) runAnsiblePlaybook(playbook string) error {
	playbookPath := filepath.Join("ansible", "playbooks", playbook+".yml")
	args := []string{
		"-i", filepath.Join(".generated", "ansible", "inventory", "hosts.yml"),
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

func (h ansibleHandler) writeHostVarsFile(hostname string, vars any) error {
	dir := filepath.Join(".generated/ansible/inventory/host_vars", hostname)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("error creating host_vars dir for %s: %v", hostname, err)
	}

	data, err := yaml.Marshal(vars)
	if err != nil {
		return fmt.Errorf("error marshaling host vars for %s: %v", hostname, err)
	}

	if err := os.WriteFile(filepath.Join(dir, "vars.yml"), data, 0o644); err != nil {
		return fmt.Errorf("error writing host vars file for %s: %v", hostname, err)
	}

	return nil
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
