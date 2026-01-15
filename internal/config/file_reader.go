package config

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"strings"

	"github.com/goccy/go-yaml"
)

var _ Reader = (*fileReader)(nil)

type fileReader struct {
	fileSystem fs.FS
	hostName   string
	verbose    bool
	paths      []string
}

// NewFileReader creates a new reader which gets key-value pairs from YML files from specified directories/files
func NewFileReader(hostName string, verbose bool, path string, opts ...func(*fileReader)) *fileReader {
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
	for _, path := range r.paths {
		info, err := fs.Stat(r.fileSystem, path)
		if err != nil {
			return nil, fmt.Errorf("error getting info for path(%s): %v", path, err)
		}

		mode := info.Mode()
		switch {
		case mode.IsDir():
			dirConfig, err := r.readDirectory(path)
			if err != nil {
				return nil, err
			}
			maps.Copy(m, dirConfig)
		case mode.IsRegular():
			fileConfig, err := r.readFile(path)
			if err != nil {
				return nil, err
			}
			maps.Copy(m, fileConfig)
		default:
			return nil, fmt.Errorf("path %s is a special file (%v) and cannot be read as config", path, mode)
		}
	}
	return NewSimpleReadResult(m), nil
}

func (r *fileReader) readDirectory(dir string) (map[string]string, error) {
	dirConfig := make(map[string]string)

	if err := fs.WalkDir(r.fileSystem, dir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".yml") && !strings.HasSuffix(path, ".yaml") {
			return nil
		}

		fileConfig, err := r.readFile(path)
		if err != nil {
			return fmt.Errorf("error reading config file(%s): %v", path, err)
		}

		maps.Copy(dirConfig, fileConfig)

		return nil
	}); err != nil {
		return nil, fmt.Errorf("error walking config directory(%s): %v", dir, err)
	}

	return dirConfig, nil
}

func (r *fileReader) readFile(file string) (map[string]string, error) {
	fileConfig := make(map[string]string)
	data, err := fs.ReadFile(r.fileSystem, file)
	if errors.Is(err, os.ErrNotExist) {
		if r.verbose {
			fmt.Fprintf(os.Stderr, "warning: %s config file not found\n", file)
		}
		return fileConfig, nil
	} else if err != nil {
		return nil, fmt.Errorf("error reading config file(%s): %v", file, err)
	}
	if err := yaml.Unmarshal(data, &fileConfig); err != nil {
		return nil, fmt.Errorf("error unmarshalling config file (%s): %v", file, err)
	}
	return fileConfig, nil
}

// WithFileSystem allows specifying a custom file system for the fileReader
// By default, it uses the OS file system, os.DirFS(".")
func WithFileSystem(fileSystem fs.FS) func(*fileReader) {
	return func(fileReader *fileReader) {
		fileReader.fileSystem = fileSystem
	}
}

// WithPath allows specifying an additional path (file or directory) to read config from
func WithPath(path string) func(*fileReader) {
	return func(fileReader *fileReader) {
		fileReader.paths = append(fileReader.paths, path)
	}
}
