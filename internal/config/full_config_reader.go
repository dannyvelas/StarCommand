package config

import (
	"errors"
	"fmt"

	"github.com/dannyvelas/homelab/internal/helpers"
)

var hostToConfig = map[string]config{
	"proxmox": newProxmoxConfig(),
}

var _ reader = (*fullConfigReader)(nil)

type fullConfigReader struct {
	hostName string
	verbose  bool
}

func NewFullConfig(hostName string, verbose bool) *fullConfigReader {
	return &fullConfigReader{
		hostName: hostName,
		verbose:  verbose,
	}
}

func (r *fullConfigReader) ReadValidated() (map[string]string, error) {
	hostConfig := hostToConfig[r.hostName]

	diagnosticMap, err := UnmarshalInto(r, hostConfig)
	if err != nil && !errors.Is(err, ErrInvalidFields) {
		return nil, fmt.Errorf("error reading host config into struct: %v", err)
	}

	results, err := validateConfig(hostConfig)
	if errors.Is(err, ErrInvalidFields) {
		return nil, fmt.Errorf("invalid or missing fields:\n%s", diagnosticMapToTable(helpers.MergeMaps(diagnosticMap, results)))
	} else if err != nil {
		return nil, fmt.Errorf("error validating config: %v", err)
	}

	if fillableConfig, ok := hostConfig.(fillableConfig); ok {
		if err := fillableConfig.FillInKeys(); err != nil {
			return nil, fmt.Errorf("error filling in fields: %v", err)
		}
	}

	configMap, err := helpers.ToMap(hostConfig)
	if err != nil {
		return nil, fmt.Errorf("error transforming host config into config map: %v", err)
	}

	return configMap, nil
}

func (r *fullConfigReader) read() (readResult, error) {
	// TODO: make this dynamic
	usingBitwarden := true

	configMap := make(map[string]string)

	// read files
	if _, err := UnmarshalInto(newFileReader(r.hostName, r.verbose), &configMap); err != nil {
		return nil, fmt.Errorf("error unmarshalling files to map: %v", err)
	}

	// read env
	if _, err := UnmarshalInto(newEnvReader(), &configMap); err != nil {
		return nil, fmt.Errorf("error unmarshalling env to map: %v", err)
	}

	if usingBitwarden {
		bitwardenSecretReader := newBitwardenSecretReader(configMap)
		diagnosticMap, err := UnmarshalInto(bitwardenSecretReader, &configMap)
		if err != nil && !errors.Is(err, ErrInvalidFields) {
			return nil, fmt.Errorf("error unmarshalling bitwarden secrets to map: %v", err)
		}

		return diagnosticReadResult{configMap: configMap, diagnosticMap: diagnosticMap}, err
	}

	return simpleReadResult{configMap: configMap}, nil
}

func (r *fullConfigReader) DryRun() (string, error) {
	hostConfig := hostToConfig[r.hostName]

	diagnosticMap, err := UnmarshalInto(r, hostConfig)
	if err != nil && !errors.Is(err, ErrInvalidFields) {
		return "", fmt.Errorf("error reading host config into struct: %v", err)
	}

	results, err := validateConfig(hostConfig)
	if err != nil && !errors.Is(err, ErrInvalidFields) {
		return "", fmt.Errorf("error validating config: %v", err)
	}

	return diagnosticMapToTable(helpers.MergeMaps(diagnosticMap, results)), nil
}
