package config

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

const configDir = "./config"

var fallbackConfigFile = filepath.Join(configDir, "all.yml")

var _ Reader = fileReader{}

type fileReader struct {
	fileSystem fs.FS
	hostName   string
	verbose    bool
}

func newFileReader(fileSystem fs.FS, hostName string, verbose bool) fileReader {
	return fileReader{
		fileSystem: fileSystem,
		hostName:   hostName,
		verbose:    verbose,
	}
}

func (r fileReader) read() (readResult, error) {
	m := make(map[string]string)
	hostConfigFile := filepath.Join(configDir, fmt.Sprintf("%s.yml", r.hostName))
	for _, file := range []string{fallbackConfigFile, hostConfigFile} {
		tempMap := make(map[string]string)
		data, err := fs.ReadFile(r.fileSystem, file)
		if errors.Is(err, os.ErrNotExist) {
			if r.verbose {
				fmt.Fprintf(os.Stderr, "warning: %s config file not found\n", file)
			}
			continue
		} else if err != nil {
			return nil, fmt.Errorf("error reading config file(%s): %v", file, err)
		}
		if err := yaml.Unmarshal(data, &tempMap); err != nil {
			return nil, fmt.Errorf("error unmarshalling config file (%s): %v", file, err)
		}
		maps.Copy(m, tempMap)
	}
	return simpleReadResult{configMap: m}, nil
}
