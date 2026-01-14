package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/dannyvelas/homelab/internal/helpers"
)

var hostToConfig = map[string]config{
	"proxmox": newProxmoxConfig(),
}

var _ Reader = (*fullConfigReader)(nil)

type fullConfigReader struct {
	fileSystem fs.FS
	environ    []string
	hostName   string
	verbose    bool
}

func NewFullConfigReader(hostName string, verbose bool, opts ...func(*fullConfigReader)) *fullConfigReader {
	fullConfigReader := &fullConfigReader{
		fileSystem: os.DirFS("."),
		environ:    os.Environ(),
		hostName:   hostName,
		verbose:    verbose,
	}

	for _, opt := range opts {
		opt(fullConfigReader)
	}

	return fullConfigReader
}

func (r *fullConfigReader) Read() (config, error) {
	hostConfig := hostToConfig[r.hostName]

	diagnosticMapInternalConfig, err := UnmarshalInto(r, hostConfig)
	if err != nil && !errors.Is(err, ErrInvalidFields) {
		return nil, fmt.Errorf("error reading host config into struct: %v", err)
	}

	diagnosticMapHostConfig, err := validateConfig(hostConfig)
	if errors.Is(err, ErrInvalidFields) {
		return nil, fmt.Errorf("invalid or missing fields:\n%s", diagnosticMapToTable(helpers.MergeMaps(diagnosticMapInternalConfig, diagnosticMapHostConfig)))
	} else if err != nil {
		return nil, fmt.Errorf("error validating config: %v", err)
	}

	if fillableConfig, ok := hostConfig.(fillableConfig); ok {
		if err := fillableConfig.FillInKeys(); err != nil {
			return nil, fmt.Errorf("error filling in fields: %v", err)
		}
	}

	return hostConfig, nil
}

func (r *fullConfigReader) read() (readResult, error) {
	// TODO: make this dynamic
	usingBitwarden := true

	configMap := make(map[string]string)

	// read files
	if _, err := UnmarshalInto(newFileReader(r.fileSystem, r.hostName, r.verbose), &configMap); err != nil {
		return nil, fmt.Errorf("error unmarshalling files to map: %v", err)
	}

	// read env
	if _, err := UnmarshalInto(newEnvReader(r.environ), &configMap); err != nil {
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

	diagnosticMapInternalConfig, err := UnmarshalInto(r, hostConfig)
	if err != nil && !errors.Is(err, ErrInvalidFields) {
		return "", fmt.Errorf("error reading host config into struct: %v", err)
	}

	diagnosticMapHostConfig, err := validateConfig(hostConfig)
	if err != nil && !errors.Is(err, ErrInvalidFields) {
		return "", fmt.Errorf("error validating config: %v", err)
	}

	return diagnosticMapToTable(helpers.MergeMaps(diagnosticMapInternalConfig, diagnosticMapHostConfig)), nil
}

func WithFilesystem(fileSystem fs.FS) func(*fullConfigReader) {
	return func(fullConfigReader *fullConfigReader) {
		fullConfigReader.fileSystem = fileSystem
	}
}

func WithEnviron(environ []string) func(*fullConfigReader) {
	return func(fullConfigReader *fullConfigReader) {
		fullConfigReader.environ = environ
	}
}
