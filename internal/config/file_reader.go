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

var _ Reader = (*fileReader)(nil)

type fileReader struct {
	fileSystem fs.FS
	hostName   string
	verbose    bool
}

// NewFileReader creates a new reader which gets key-value pairs from YML files from specified directories/files
func NewFileReader(hostName string, verbose bool, opts ...func(*fileReader)) *fileReader {
	fileReader := fileReader{
		fileSystem: os.DirFS("."),
		hostName:   hostName,
		verbose:    verbose,
	}

	for _, opt := range opts {
		opt(&fileReader)
	}

	return &fileReader
}

func (r *fileReader) Read() (ReadResult, error) {
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
	return NewSimpleReadResult(m), nil
}

// WithFileSystem allows specifying a custom file system for the fileReader
// By default, it uses the OS file system, os.DirFS(".")
func WithFileSystem(fileSystem fs.FS) func(*fileReader) {
	return func(fileReader *fileReader) {
		fileReader.fileSystem = fileSystem
	}
}
