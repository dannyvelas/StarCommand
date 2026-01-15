package config

import (
	"fmt"
	"io/fs"
	"os"
)

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

func (r *fullConfigReader) read() (readResult, error) {
	configMap := make(map[string]string)

	// read files
	if _, err := Unmarshal(newFileReader(r.fileSystem, r.hostName, r.verbose), &configMap); err != nil {
		return nil, fmt.Errorf("error unmarshalling files to map: %v", err)
	}

	// read env
	if _, err := Unmarshal(newEnvReader(r.environ), &configMap); err != nil {
		return nil, fmt.Errorf("error unmarshalling env to map: %v", err)
	}

	bitwardenSecretReader := newBitwardenSecretReader(configMap)
	diagnosticMap, err := Unmarshal(bitwardenSecretReader, &configMap)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling bitwarden secrets to map: %v", err)
	}

	return diagnosticReadResult{configMap: configMap, diagnosticMap: diagnosticMap}, err
}

//func (r *fullConfigReader) DryRun() (string, error) {
//	diagnosticMap, err := UnmarshalIntoStruct(r, hostConfig)
//	if err != nil && !errors.Is(err, ErrInvalidFields) {
//		return "", fmt.Errorf("error reading host config into struct: %v", err)
//	}
//
//	return diagnosticMapToTable(diagnosticMap), nil
//}

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
