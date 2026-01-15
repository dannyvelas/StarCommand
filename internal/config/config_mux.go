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

func NewConfigMux(hostName string, verbose bool, opts ...func(*configMux)) *configMux {
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
		reader := readerFn(configMap)
		readerDiagnostics, err := Unmarshal(reader, &configMap)
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling bitwarden secrets to map: %v", err)
		}
		maps.Copy(allDiagnostics, readerDiagnostics)
	}

	return diagnosticReadResult{configMap: configMap, diagnostics: allDiagnostics}, nil
}

func WithFileReader(opts ...func(*fileReader)) func(*configMux) {
	return func(configMux *configMux) {
		configMux.readerFns = append(
			configMux.readerFns,
			func(_ map[string]string) Reader {
				return NewFileReader(configMux.hostName, configMux.verbose, opts...)
			},
		)
	}
}

func WithEnvReader(opts ...func(*envReader)) func(*configMux) {
	return func(configMux *configMux) {
		configMux.readerFns = append(
			configMux.readerFns,
			func(_ map[string]string) Reader {
				return NewEnvReader(opts...)
			},
		)
	}
}

func WithBitwardenSecretReader() func(*configMux) {
	return func(configMux *configMux) {
		configMux.readerFns = append(configMux.readerFns, func(configMap map[string]string) Reader {
			return NewBitwardenSecretReader(configMap)
		})
	}
}
