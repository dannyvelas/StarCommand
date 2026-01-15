package config

import (
	"fmt"
	"maps"
)

var _ Reader = (*fullConfigReader)(nil)

type fullConfigReader struct {
	hostName  string
	verbose   bool
	readerFns []func(configMap map[string]string) Reader
}

func NewFullConfigReader(hostName string, verbose bool, opts ...func(*fullConfigReader)) *fullConfigReader {
	fullConfigReader := &fullConfigReader{
		hostName: hostName,
		verbose:  verbose,
	}

	for _, opt := range opts {
		opt(fullConfigReader)
	}

	return fullConfigReader
}

func (r *fullConfigReader) read() (readResult, error) {
	configMap, allDiagnostics := make(map[string]string), make(map[string]string)
	for _, readerFn := range r.readerFns {
		Reader := readerFn(configMap)
		readerDiagnosticMap, err := Unmarshal(Reader, &configMap)
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling bitwarden secrets to map: %v", err)
		}
		maps.Copy(allDiagnostics, readerDiagnosticMap)
	}

	return diagnosticReadResult{configMap: configMap, diagnosticMap: allDiagnostics}, nil
}

func WithReader(r Reader) func(*fullConfigReader) {
	return func(fullConfigReader *fullConfigReader) {
		fullConfigReader.readerFns = append(
			fullConfigReader.readerFns,
			func(_ map[string]string) Reader { return r },
		)
	}
}

func WithLazyReader(fn func(configMap map[string]string) Reader) func(*fullConfigReader) {
	return func(fullConfigReader *fullConfigReader) {
		fullConfigReader.readerFns = append(fullConfigReader.readerFns, fn)
	}
}
