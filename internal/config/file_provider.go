package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

var (
	_ provider          = fileProvider{}
	_ unvalidatedReader = fileProvider{}
)

type fileProvider struct {
	hostName string
	verbose  bool
}

func newFileProvider(hostName string, verbose bool) fileProvider {
	return fileProvider{
		hostName: hostName,
		verbose:  verbose,
	}
}

func (p fileProvider) UnmarshalInto(target any) error {
	fileMap, err := p.ReadUnvalidated()
	if err != nil {
		return fmt.Errorf("error reading file configs: %v", err)
	}

	// decode files
	if err := decode(fileMap, target); err != nil {
		return fmt.Errorf("error decoding file configs into a map: %v", err)
	}

	return nil
}

func (p fileProvider) ReadUnvalidated() (map[string]string, error) {
	m := make(map[string]string)
	hostConfigFile := filepath.Join(configDir, fmt.Sprintf("%s.yml", p.hostName))
	for _, file := range []string{fallbackConfigFile, hostConfigFile} {
		data, err := os.ReadFile(file)
		if errors.Is(err, os.ErrNotExist) {
			if p.verbose {
				fmt.Fprintf(os.Stderr, "warning: %s config file not found\n", file)
			}
			continue
		} else if err != nil {
			return nil, fmt.Errorf("error reading config file(%s): %v", file, err)
		}
		if err := yaml.Unmarshal(data, m); err != nil {
			return nil, fmt.Errorf("error unmarshalling config file (%s): %v", file, err)
		}
	}
	return m, nil
}
