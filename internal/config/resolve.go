package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dannyvelas/homelab/internal/client"
	"github.com/goccy/go-yaml"
)

const configDir = "./config"

var fallbackConfigFile = filepath.Join(configDir, "all.yml")

var hostToConfig = map[string]config{
	"proxmox": newProxmoxConfig(),
}

func Resolve(hostName string, verbose bool) (map[string]string, error) {
	hostConfig, err := readConfigs(hostName, verbose)
	if err != nil {
		return nil, fmt.Errorf("error reading configs: %v", err)
	}

	validateResult, ok, err := validateStruct(hostConfig)
	if err != nil {
		return nil, fmt.Errorf("error validating config: %v", err)
	} else if !ok {
		return nil, fmt.Errorf("error: invalid configs: %s", fmtTable(validateResult))
	}

	if fillableConfig, ok := hostConfig.(fillableConfig); ok {
		if err := fillableConfig.FillInKeys(); err != nil {
			return nil, fmt.Errorf("error filling in fields: %v", err)
		}
	}

	m, err := configToMap(hostConfig)
	if err != nil {
		return nil, fmt.Errorf("error transforming config to map: %v", err)
	}

	return m, nil
}

func DryRun(hostName string, verbose bool) (string, error) {
	hostConfig, err := readConfigs(hostName, verbose)
	if err != nil {
		return "", fmt.Errorf("error reading configs: %v", err)
	}

	validateResult, _, err := validateStruct(hostConfig)
	if err != nil {
		return "", fmt.Errorf("error validating config: %v", err)
	}

	return fmtTable(validateResult), nil
}

func readConfigs(hostName string, verbose bool) (config, error) {
	hostConfig, ok := hostToConfig[hostName]
	if !ok {
		return nil, fmt.Errorf("unrecognized host: %s", hostName)
	}

	configMap := make(map[string]string)
	if err := filesUnmarshalInto(hostName, verbose, configMap); err != nil {
		return nil, fmt.Errorf("error reading file configs: %v", err)
	}

	if err := envUnmarshalInto(configMap); err != nil {
		return nil, fmt.Errorf("error reading env into map: %v", err)
	}

	bitwardenConfig := newBitwardenConfig()
	if err := decode(configMap, &bitwardenConfig); err != nil {
		return nil, fmt.Errorf("error reading bitwarden config into a map: %v", err)
	}

	results, ok, err := validateStruct(bitwardenConfig)
	if err != nil {
		return nil, fmt.Errorf("error validating bitwarden config: %v", err)
	} else if !ok {
		return nil, fmt.Errorf("error: invalid bitwarden configs: %s", fmtTable(results))
	}

	bwClient, err := client.NewBitwardenClient(
		bitwardenConfig.APIURL,
		bitwardenConfig.IdentityURL,
		bitwardenConfig.AccessToken,
		bitwardenConfig.OrganizationID,
		bitwardenConfig.ProjectID,
		bitwardenConfig.StateFilePath,
	)
	if err != nil {
		return nil, fmt.Errorf("error initializing bitwarden client: %v", err)
	}

	if err := bwClient.UnmarshalInto(configMap); err != nil {
		return nil, fmt.Errorf("error filling host config struct with bitwarden secrets: %v", err)
	}

	if err := decode(configMap, hostConfig); err != nil {
		return nil, fmt.Errorf("error unmarshalling all configs into host config: %v", err)
	}

	return hostConfig, nil
}

func filesUnmarshalInto(hostName string, verbose bool, target map[string]string) error {
	hostConfigFile := filepath.Join(configDir, fmt.Sprintf("%s.yml", hostName))
	for _, file := range []string{fallbackConfigFile, hostConfigFile} {
		data, err := os.ReadFile(file)
		if errors.Is(err, os.ErrNotExist) {
			if verbose {
				fmt.Fprintf(os.Stderr, "warning: %s config file not found\n", file)
			}
			continue
		} else if err != nil {
			return fmt.Errorf("error reading config file(%s): %v", file, err)
		}
		if err := yaml.Unmarshal(data, target); err != nil {
			return fmt.Errorf("error unmarshalling config file (%s): %v", file, err)
		}
	}
	return nil
}

func envUnmarshalInto(target map[string]string) error {
	environ := os.Environ()
	for _, entry := range environ {
		if entry != "" {
			key, value, err := split(entry)
			if err != nil {
				return fmt.Errorf("error splitting: %v", err)
			}
			target[key] = value
		}
	}
	return nil
}

func split(entry string) (string, string, error) {
	parts := strings.SplitN(entry, "=", 2)
	switch len(parts) {
	case 0:
		return "", "", fmt.Errorf("cannot split empty string")
	case 1:
		return parts[0], "", nil
	default:
		return parts[0], parts[1], nil
	}
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
