package config

import (
	"fmt"
	"path/filepath"
)

const configDir = "./config"

var fallbackConfigFile = filepath.Join(configDir, "all.yml")

var hostToConfig = map[string]config{
	"proxmox": newProxmoxConfig(),
}

var (
	_ validatedReader   = fullConfig{}
	_ unvalidatedReader = fullConfig{}
)

type fullConfig struct {
	hostName string
	verbose  bool
}

func NewFullConfig(hostName string, verbose bool) fullConfig {
	return fullConfig{
		hostName: hostName,
		verbose:  verbose,
	}
}

func (p fullConfig) ReadValidated() (map[string]string, error) {
	configMap, err := p.ReadUnvalidated()
	if err != nil {
		return nil, fmt.Errorf("error reading configs: %v", err)
	}

	hostConfig := hostToConfig[p.hostName]
	if err := decode(configMap, hostConfig); err != nil {
		return nil, fmt.Errorf("error transforming map to host config: %v", err)
	}

	validateResult, ok, err := validateConfig(hostConfig)
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

	return configMap, nil
}

func (p fullConfig) ReadUnvalidated() (map[string]string, error) {
	// TODO: make this dynamic
	usingBitwarden := true

	configMap := make(map[string]string)

	// read files
	if err := UnmarshalInto(newFileReader(p.hostName, p.verbose), configMap); err != nil {
		return nil, fmt.Errorf("error unmarshalling files to map: %v", err)
	}

	// read env
	if err := UnmarshalInto(newEnvReader(), configMap); err != nil {
		return nil, fmt.Errorf("error unmarshalling env to map: %v", err)
	}

	if usingBitwarden {
		if err := UnmarshalInto(newBitwardenSecretReader(configMap), configMap); err != nil {
			return nil, fmt.Errorf("error unmarshalling bitwarden secrets to map: %v", err)
		}
	}

	return configMap, nil
}

func (p fullConfig) DryRun(hostName string, verbose bool) (string, error) {
	hostConfig, err := p.ReadUnvalidated()
	if err != nil {
		return "", fmt.Errorf("error reading configs: %v", err)
	}

	validateResult, _, err := validateConfig(hostConfig)
	if err != nil {
		return "", fmt.Errorf("error validating config: %v", err)
	}

	return fmtTable(validateResult), nil
}
