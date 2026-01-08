package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dannyvelas/homelab/internal/client"
	"github.com/dannyvelas/homelab/internal/helpers"
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

	validateResult, ok, err := hostConfig.Validate()
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

	validateResult, _, err := hostConfig.Validate()
	if err != nil {
		return "", fmt.Errorf("error validating config: %v", err)
	}

	return fmtTable(validateResult), nil
}

func readConfigs(hostName string, verbose bool) (config, error) {
	bitwardenConfig := defaultBitwardenConfig

	hostConfig, ok := hostToConfig[hostName]
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
		if err := yaml.Unmarshal(data, &bitwardenConfig); err != nil {
			return nil, fmt.Errorf("error unmarshalling config file (%s) to root config: %v", file, err)
		}
		if err := yaml.Unmarshal(data, hostConfig); err != nil {
			return nil, fmt.Errorf("error unmarshalling config file (%s) to host config: %v", file, err)
		}
	}

	if missingBitwardenFields := getMissingBitwardenFields(); len(missingBitwardenFields) > 0 {
		fmt.Fprintf(os.Stderr, "warning: some bitwarden ENV variables are missing:%s\nBitwarden secrets will not be loaded.\n", helpers.StringSliceToBulletedList(missingBitwardenFields))
		return hostConfig, nil
	}

	bwClient, err := client.NewBitwardenClient(
		bitwardenConfig.BitwardenAPIURL,
		bitwardenConfig.BitwardenIdentityURL,
		"",
		"",
		"",
		"",
	)
	if err != nil {
		return nil, fmt.Errorf("error initializing bitwarden client: %v", err)
	}

	if err := bwClient.FillStruct(hostConfig); err != nil {
		return nil, fmt.Errorf("error filling host config struct with bitwarden secrets: %v", err)
	}

	return hostConfig, nil
}

func getMissingBitwardenFields() []string {
	missing := make([]string, 0)

	//if env.BitwardenAccessToken == "" {
	//	missing = append(missing, "BWS_ACCESS_TOKEN")
	//}

	//if env.BitwardenProjectID == "" {
	//	missing = append(missing, "BWS_PROJECT_ID")
	//}

	//if env.BitwardenOrganizationID == "" {
	//	missing = append(missing, "BWS_ORGANIZATION_ID")
	//}

	return missing
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
