package config

import (
	"fmt"
)

var hostToConfig = map[string]config{
	"proxmox": newProxmoxConfig(),
}

var (
	_ validatedReader   = fullConfigReader{}
	_ unvalidatedReader = fullConfigReader{}
)

type fullConfigReader struct {
	hostName string
	verbose  bool
}

func NewFullConfig(hostName string, verbose bool) fullConfigReader {
	return fullConfigReader{
		hostName: hostName,
		verbose:  verbose,
	}
}

func (p fullConfigReader) ReadValidated() (map[string]string, error) {
	hostConfig := hostToConfig[p.hostName]
	if err := UnmarshalInto(p, hostConfig); err != nil {
		return nil, fmt.Errorf("error reading host config into struct: %v", err)
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

	configMap := make(map[string]string)
	if err := decode(hostConfig, configMap); err != nil {
		return nil, fmt.Errorf("error transforming host config into config map: %v", err)
	}

	return configMap, nil
}

func (p fullConfigReader) ReadUnvalidated() (map[string]string, error) {
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

func (p fullConfigReader) DryRun() (string, error) {
	hostConfig := hostToConfig[p.hostName]
	if err := UnmarshalInto(p, hostConfig); err != nil {
		return "", fmt.Errorf("error reading host config into struct: %v", err)
	}

	validateResult, _, err := validateConfig(hostConfig)
	if err != nil {
		return "", fmt.Errorf("error validating config: %v", err)
	}

	return fmtTable(validateResult), nil
}
