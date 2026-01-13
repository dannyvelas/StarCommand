package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

const configDir = "./config"

var fallbackConfigFile = filepath.Join(configDir, "all.yml")

var _ unvalidatedReader = fileReader{}

type fileReader struct {
	hostName string
	verbose  bool
}

func newFileReader(hostName string, verbose bool) fileReader {
	return fileReader{
		hostName: hostName,
		verbose:  verbose,
	}
}

func (p fileReader) ReadUnvalidated() (unvalidatedResult, error) {
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
		if err := yaml.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("error unmarshalling config file (%s): %v", file, err)
		}
	}
	return simpleUnvalidatedResult{configMap: m}, nil
}
