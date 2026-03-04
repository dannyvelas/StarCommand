package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dannyvelas/starcommand/config"
	"github.com/dannyvelas/starcommand/internal/helpers"
	"github.com/goccy/go-yaml"
	"golang.org/x/crypto/ssh"
)

type bootstrapHostEntry struct {
	Name                 string
	IP                   string
	SSH                  config.SSHConfig
	AutoUpdateRebootTime string
}

type bootstrapHostVars struct {
	AnsibleHost          string `yaml:"ansible_host"`
	AnsiblePort          string `yaml:"ansible_port"`
	AnsibleSSHPrivateKey string `yaml:"ansible_ssh_private_key_file"`
	AnsibleUser          string `yaml:"ansible_user"`
	SSHPublicKey         string `yaml:"ssh_public_key"`
	AutoUpdateRebootTime string `yaml:"auto_update_reboot_time"`
}

type ansibleBootstrapConfig struct {
	Hosts []bootstrapHostEntry `json:"-" required:"true"`

	// Sensitive
	AdminEmail    string `json:"admin_email" sensitive:"true" prompt:"Admin email"`
	AdminPassword string `json:"admin_password" sensitive:"true" prompt:"Admin password"`
}

func newAnsibleBootstrapConfig(c *config.Config) *ansibleBootstrapConfig {
	bootstrapConfig := new(ansibleBootstrapConfig)
	for _, host := range c.Hosts {
		bootstrapConfig.Hosts = append(bootstrapConfig.Hosts, bootstrapHostEntry{
			Name:                 host.Name,
			IP:                   host.IP,
			SSH:                  host.SSH,
			AutoUpdateRebootTime: host.AutoUpdateRebootTime,
		})
	}
	return bootstrapConfig
}

func (c *ansibleBootstrapConfig) generateHostVars() error {
	for _, host := range c.Hosts {
		ansibleUser, err := determineAnsibleUser(host.SSH.User, host.IP, portToString(host.SSH.Port), host.SSH.PrivateKeyPath)
		if err != nil {
			return fmt.Errorf("error determining ansible user for %s: %v", host.Name, err)
		}

		expandedPrivateKey, err := helpers.ExpandPath(host.SSH.PrivateKeyPath)
		if err != nil {
			return fmt.Errorf("error expanding private key path for %s: %v", host.Name, err)
		}

		expandedPublicKey, err := helpers.ExpandPath(host.SSH.PublicKeyPath)
		if err != nil {
			return fmt.Errorf("error expanding public key path for %s: %v", host.Name, err)
		}

		pubKeyBytes, err := os.ReadFile(expandedPublicKey)
		if err != nil {
			return fmt.Errorf("error reading public key for %s: %v", host.Name, err)
		}

		autoUpdateRebootTime := host.AutoUpdateRebootTime
		if autoUpdateRebootTime == "" {
			autoUpdateRebootTime = "05:00"
		}

		vars := bootstrapHostVars{
			AnsibleHost:          host.IP,
			AnsiblePort:          portToString(host.SSH.Port),
			AnsibleSSHPrivateKey: expandedPrivateKey,
			AnsibleUser:          ansibleUser,
			SSHPublicKey:         string(pubKeyBytes),
			AutoUpdateRebootTime: autoUpdateRebootTime,
		}

		dir := filepath.Join(".generated", "host_vars", host.Name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("error creating host_vars dir for %s: %v", host.Name, err)
		}

		data, err := yaml.Marshal(vars)
		if err != nil {
			return fmt.Errorf("error marshaling host vars for %s: %v", host.Name, err)
		}

		if err := os.WriteFile(filepath.Join(dir, "vars.yml"), data, 0o644); err != nil {
			return fmt.Errorf("error writing host vars file for %s: %v", host.Name, err)
		}
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
