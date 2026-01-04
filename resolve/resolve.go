package resolve

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

const configDir = "./config"

var fallbackConfigFile = filepath.Join(configDir, "all.yml")

const (
	keySSHPublicKeyPath     = "ssh_public_key_path"
	keyNodeCIDRAddress      = "node_cidr_address"
	keyGatewayAddress       = "gateway_address"
	keyPhysicalNIC          = "physical_nic"
	keyVaultAdminPassword   = "vault_admin_password"
	keySSHPort              = "ssh_port"
	keyAutoUpdateRebootTime = "auto_update_reboot_time"
	keyAdminEmail           = "admin_email"
	keySMTPUser             = "smtp_user"
	keySMTPPassword         = "smtp_password"
)

var hostRequiredKeys = map[string][]string{
	"proxmox": {
		keySSHPublicKeyPath,
		keyNodeCIDRAddress,
		keyGatewayAddress,
		keyPhysicalNIC,
		keyVaultAdminPassword,
		keySSHPort,
		keyAutoUpdateRebootTime,
		keyAdminEmail,
		keySMTPUser,
		keySMTPPassword,
	},
}

func ResolveConfig(verbose bool, hostName string) (map[string]string, error) {
	conf := make(map[string]string)

	hostConfigFile := filepath.Join(configDir, fmt.Sprintf("%s.yml", hostName))
	for _, file := range []string{fallbackConfigFile, hostConfigFile} {
		data, err := os.ReadFile(file)
		if errors.Is(err, os.ErrNotExist) {
			if verbose {
				fmt.Fprintf(os.Stderr, "warning: %s config file not found\n", file)
			}
			continue
		} else if err != nil {
			return nil, fmt.Errorf("error reading config file(%s): %v", file, err)
		}
		if err := yaml.Unmarshal(data, &conf); err != nil {
			return nil, fmt.Errorf("error unmarshalling config file (%s): %v", file, err)
		}
	}

	// validate whether all necessary configs are present
	requiredKeys := hostRequiredKeys[hostName]
	missingKeys := getMissingKeys(conf, requiredKeys)
	if len(missingKeys) > 0 {
		return nil, fmt.Errorf("error: missing values for the following keys: %v", missingKeys)
	}

	conf["node_ip"] = strings.Split(conf["node_cidr_address"], "/")[0]
	return conf, nil
}

func getMissingKeys(conf map[string]string, requiredKeys []string) []string {
	missingKeys := make([]string, 0)
	for _, requiredKey := range requiredKeys {
		if _, ok := conf[requiredKey]; !ok {
			missingKeys = append(missingKeys, requiredKey)
		}
	}
	return missingKeys
}
