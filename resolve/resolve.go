package resolve

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dannyvelas/homelab/client"
	"github.com/dannyvelas/homelab/env"
	"github.com/dannyvelas/homelab/helpers"
	"gopkg.in/yaml.v2"
)

const configDir = "./config"

var fallbackConfigFile = filepath.Join(configDir, "all.yml")

var hostRequiredKeys = map[string]Config{
	"proxmox": NewProxmoxConfig(),
}

func ResolveConfig(env env.Env, verbose bool, hostName string) (map[string]string, error) {
	rootConfig := defaultRootConfig

	hostConfig, ok := hostRequiredKeys[hostName]
	if !ok {
		return nil, fmt.Errorf("unrecognized host: %s", hostName)
	}

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
		if err := yaml.Unmarshal(data, &rootConfig); err != nil {
			return nil, fmt.Errorf("error unmarshalling config file (%s) to root config: %v", file, err)
		}
		if err := yaml.Unmarshal(data, hostConfig); err != nil {
			return nil, fmt.Errorf("error unmarshalling config file (%s) to host config: %v", file, err)
		}
	}

	bwClient, err := client.NewBitwardenClient(
		rootConfig.BitwardenAPIURL,
		rootConfig.BitwardenIdentityURL,
		env.BitwardenAccessToken,
		env.BitwardenOrganizationID,
		env.BitwardenProjectID,
		env.BitwardenStateFilePath,
	)
	if err != nil {
		return nil, fmt.Errorf("error initializing bitwarden client: %v", err)
	}

	if err := bwClient.FillStruct(hostConfig); err != nil {
		return nil, fmt.Errorf("error filling host config struct with bitwarden secrets: %v", err)
	}

	configErrors := hostConfig.Validate()
	if len(configErrors) > 0 {
		return nil, fmt.Errorf("error: invalid configs: %s", helpers.MapToBulletedList(configErrors))
	}

	if err := hostConfig.FillInKeys(); err != nil {
		return nil, fmt.Errorf("error filling in fields: %v", err)
	}

	m, err := configToMap(hostConfig)
	if err != nil {
		return nil, fmt.Errorf("error transforming config to map: %v", err)
	}

	return m, nil
}

func configToMap(c any) (map[string]string, error) {
	bytes, err := json.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("error marshalling config: %v", err)
	}

	var m map[string]string
	if err := json.Unmarshal(bytes, &m); err != nil {
		return nil, fmt.Errorf("error unmarshalling config to map: %v", err)
	}

	return m, nil
}
