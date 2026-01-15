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

// NewConfigMux creates a new config mux which can read from multiple readers
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

func (r *configMux) Read() (ReadResult, error) {
	configMap, allDiagnostics := make(map[string]string), make(map[string]string)
	for _, readerFn := range r.readerFns {
		reader := readerFn(configMap)
		readerDiagnostics, err := Unmarshal(reader, &configMap)
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling bitwarden secrets to map: %v", err)
		}
		maps.Copy(allDiagnostics, readerDiagnostics)
	}

	return NewDiagnosticReadResult(configMap, allDiagnostics), nil
}

// WithFileReader adds a file reader to the config mux
func WithFileReader(path string, opts ...func(*fileReader)) func(*configMux) {
	return func(configMux *configMux) {
		configMux.readerFns = append(
			configMux.readerFns,
			func(_ map[string]string) Reader {
				return NewFileReader(configMux.hostName, configMux.verbose, path, opts...)
			},
		)
	}
}

// WithEnvReader adds an environment variable reader to the config mux
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

// WithBitwardenSecretReader adds a Bitwarden secret reader to the config mux
func WithBitwardenSecretReader() func(*configMux) {
	return func(configMux *configMux) {
		configMux.readerFns = append(configMux.readerFns, func(configMap map[string]string) Reader {
			return NewBitwardenSecretReader(configMap)
		})
	}
}

// WithCustomReader lets you add your own custom reader to the mux
// your custom reader just needs to implement the "Reader" interface
// The difference between WithCustomReader and WithCustomLazyReader is:
// - WithCustomReader asks for an already-initialized reader
// - WithCustomLazyReader asks for a function to initialize a reader
// This function is useful if your reader can be initialized at
// the same time as the mux.
// WithCustomLazyReader is more powerful, but WithCustomReader is
// simpler to use and syntactically terse
func WithCustomReader(r Reader) func(*configMux) {
	return func(configMux *configMux) {
		configMux.readerFns = append(
			configMux.readerFns,
			func(_ map[string]string) Reader { return r },
		)
	}
}

// WithCustomLazyReader lets you add your own custom reader to the mux
// The difference between WithCustomReader and WithCustomLazyReader is:
// - WithCustomReader asks for an already-initialized reader
// - WithCustomLazyReader asks for a function to initialize a reader
// This function is useful if your reader needs to be initialized after
// some config values have already been read.
// For example, the BitwardenSecretReader would need to be initialized this way
// because it expects a map of configs as an argument.
// It uses this map to try to authenticate to Bitwarden
func WithCustomLazyReader(fn func(configMap map[string]string) Reader) func(*configMux) {
	return func(configMux *configMux) {
		configMux.readerFns = append(configMux.readerFns, fn)
	}
}
