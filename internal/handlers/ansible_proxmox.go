package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/dannyvelas/homelab/internal/client"
	"golang.org/x/crypto/ssh"
)

var (
	errConnectingSSH = errors.New("error connecting via ssh")
	errAlreadyExists = errors.New("resource already exists")
)

var _ Handler = AnsibleProxmoxHandler{}

type AnsibleProxmoxHandler struct{}

func NewAnsibleProxmoxHandler() AnsibleProxmoxHandler {
	return AnsibleProxmoxHandler{}
}

func (h AnsibleProxmoxHandler) GetConfig(_ string) any {
	return newAnsibleProxmoxConfig()
}

func (h AnsibleProxmoxHandler) Execute(config any, hostAlias string) (map[string]string, error) {
	diagnostics := make(map[string]string)

	ansibleProxmoxConfig, ok := config.(*ansibleProxmoxConfig)
	if !ok {
		return diagnostics, fmt.Errorf("internal type error converting config to ansible proxmox config. found: %T", config)
	}

	if err := h.runAnsiblePlaybook(ansibleProxmoxConfig); err != nil {
		return diagnostics, fmt.Errorf("error running ansible playbook: %v", err)
	}

	token, err := h.createTokenForTerraformUser(ansibleProxmoxConfig)
	if errors.Is(err, errAlreadyExists) {
		diagnostics["Create Terraform API token"] = fmt.Sprintf("skipping: %v", errAlreadyExists)
		return diagnostics, nil
	} else if err != nil {
		return diagnostics, fmt.Errorf("error creating token for terraform user: %v", err)
	}

	if err := h.addTerraformTokenToBitwarden(ansibleProxmoxConfig, token); errors.Is(err, errAlreadyExists) {
		diagnostics["Add Terraform API token to Bitwarden"] = fmt.Sprintf("skipping: %v", errAlreadyExists)
		return diagnostics, nil
	} else if err != nil {
		return diagnostics, fmt.Errorf("error adding secret to bitwarden: %v", err)
	}

	return diagnostics, nil
}

func (h AnsibleProxmoxHandler) runAnsiblePlaybook(config *ansibleProxmoxConfig) error {
	proxmoxAddr := fmt.Sprintf("%s:%s", config.NodeIP, config.SSHPort)
	client, sshErr := h.getSSHClient(config.SSHUser, proxmoxAddr, config.SSHPrivateKeyPath)
	if sshErr != nil && !errors.Is(sshErr, errConnectingSSH) {
		return fmt.Errorf("error checking if ssh is accessible to proxmox host: %v", sshErr)
	} else if sshErr == nil {
		client.Close()
	}

	tmpFile, err := os.CreateTemp("", "labctl-vars-*.json")
	if err != nil {
		return fmt.Errorf("error creating temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

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

func (h AnsibleProxmoxHandler) createTokenForTerraformUser(config *ansibleProxmoxConfig) (string, error) {
	proxmoxAddr := fmt.Sprintf("%s:%s", config.NodeIP, config.SSHPort)
	sshClient, err := h.getSSHClient(config.SSHUser, proxmoxAddr, config.SSHPrivateKeyPath)
	if err != nil {
		return "", fmt.Errorf("error getting ssh client after running ansible: %v", err)
	}
	defer sshClient.Close()

	session, err := sshClient.NewSession()
	if err != nil {
		return "", fmt.Errorf("error creating ssh session after running ansible: %v", err)
	}

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	err = session.Run("sudo pveum user token add terraform@pve provider --privsep=0")
	if stderrString := stderr.String(); err != nil && strings.Contains(stderrString, "Token already exists") {
		return "", errAlreadyExists
	} else if err != nil {
		return "", fmt.Errorf("error creating token for terraform user in proxmox: %v", stderrString)
	}

	return stdout.String(), nil
}

func (h AnsibleProxmoxHandler) addTerraformTokenToBitwarden(config *ansibleProxmoxConfig, token string) error {
	bwClient, err := client.NewBitwardenClient(
		config.BitwardenAPIURL,
		config.BitwardenIdentityURL,
		config.BitwardenAccessToken,
		config.BitwardenOrganizationID,
		config.BitwardenProjectID,
		config.BitwardenStateFilePath,
	)
	if err != nil {
		return fmt.Errorf("error initializing bitwarden client: %v", err)
	}

	secrets, err := bwClient.ReadSecrets()
	if err != nil {
		return fmt.Errorf("error reading bitwarden secrets: %v", err)
	}

	existingSecret, ok := secrets[config.BitwardenTerraformTokenKey]
	if ok && token == existingSecret.Value {
		return errAlreadyExists
	}

	if ok && token != existingSecret.Value {
		// without this, bwClient will create an additional secret with the same key
		if err := bwClient.DeleteSecret(existingSecret.ID); err != nil {
			return fmt.Errorf("error deleting previously-existing secret with a different value: %v", err)
		}
	}

	if err := bwClient.CreateSecret(config.BitwardenTerraformTokenKey, token); err != nil {
		return fmt.Errorf("error creating bitwarden secret: %v", err)
	}

	return nil
}

func (h AnsibleProxmoxHandler) getSSHClient(user, addr, privateKeyPath string) (*ssh.Client, error) {
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
