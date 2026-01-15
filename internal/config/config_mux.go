package config

import (
	"fmt"
	"maps"
)

var _ Reader = (*configMux)(nil)

type configMux struct {
	hostName  string
	verbose   bool
	readerFns []func(configMap map[string]string) Reader
}

func NewFullConfigReader(hostName string, verbose bool, opts ...func(*configMux)) *configMux {
	configMux := &configMux{
		hostName: hostName,
		verbose:  verbose,
	}

	for _, opt := range opts {
		opt(configMux)
	}

	return configMux
}

func (r *configMux) read() (readResult, error) {
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

func WithReader(r Reader) func(*configMux) {
	return func(configMux *configMux) {
		configMux.readerFns = append(
			configMux.readerFns,
			func(_ map[string]string) Reader { return r },
		)
	}
}

func WithLazyReader(fn func(configMap map[string]string) Reader) func(*configMux) {
	return func(configMux *configMux) {
		configMux.readerFns = append(configMux.readerFns, fn)
	}
}
