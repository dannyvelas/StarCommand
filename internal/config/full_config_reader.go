package config

import (
	"errors"
	"fmt"

	"github.com/dannyvelas/homelab/internal/helpers"
)

var hostToConfig = map[string]config{
	"proxmox": newProxmoxConfig(),
}

var (
	_ unvalidatedReader = (*fullConfigReader)(nil)
	_ diagnosticReader  = (*fullConfigReader)(nil)
)

type fullConfigReader struct {
	hostName      string
	verbose       bool
	diagnosticMap map[string]string
}

func NewFullConfig(hostName string, verbose bool) *fullConfigReader {
	return &fullConfigReader{
		hostName:      hostName,
		verbose:       verbose,
		diagnosticMap: nil,
	}
}

func (p *fullConfigReader) ReadValidated() (map[string]string, error) {
	hostConfig := hostToConfig[p.hostName]

	if err := UnmarshalInto(p, hostConfig); err != nil && !errors.Is(err, ErrInvalidFields) {
		return nil, fmt.Errorf("error reading host config into struct: %v", err)
	}

	diagnosticMap := p.GetDiagnosticMap()

	results, err := validateConfig(hostConfig)
	if err != nil && !errors.Is(err, ErrInvalidFields) {
		return nil, err
	} else if errors.Is(err, ErrInvalidFields) {
		return nil, fmt.Errorf("invalid or missing fields:\n%s", diagnosticMapToTable(helpers.MergeMaps(diagnosticMap, results)))
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

func (p *fullConfigReader) ReadUnvalidated() (map[string]string, error) {
	// TODO: make this dynamic
	usingBitwarden := true

	configMap := make(map[string]string)

	// read files
	if err := UnmarshalInto(newFileReader(p.hostName, p.verbose), &configMap); err != nil {
		return nil, fmt.Errorf("error unmarshalling files to map: %v", err)
	}

	// read env
	if err := UnmarshalInto(newEnvReader(), &configMap); err != nil {
		return nil, fmt.Errorf("error unmarshalling env to map: %v", err)
	}

	if usingBitwarden {
		bitwardenSecretReader := newBitwardenSecretReader(configMap)
		err := UnmarshalInto(bitwardenSecretReader, &configMap)
		if err != nil && !errors.Is(err, ErrInvalidFields) {
			return nil, fmt.Errorf("error unmarshalling bitwarden secrets to map: %v", err)
		}

		p.diagnosticMap = bitwardenSecretReader.GetDiagnosticMap()
		if errors.Is(err, ErrInvalidFields) {
			return nil, ErrInvalidFields
		}
	}

	return configMap, nil
}

func (p *fullConfigReader) DryRun() (string, error) {
	hostConfig := hostToConfig[p.hostName]

	if err := UnmarshalInto(p, hostConfig); err != nil && !errors.Is(err, ErrInvalidFields) {
		return "", fmt.Errorf("error reading host config into struct: %v", err)
	}

	diagnosticMap := p.GetDiagnosticMap()

	results, err := validateConfig(hostConfig)
	if err != nil {
		return "", fmt.Errorf("error validating config: %v", err)
	}

	return diagnosticMapToTable(helpers.MergeMaps(diagnosticMap, results)), nil
}

func (p *fullConfigReader) GetDiagnosticMap() map[string]string {
	return p.diagnosticMap
}
