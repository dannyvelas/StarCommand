package config

import (
	"fmt"
	"os"
	"strings"
)

var _ Reader = envReader{}

type envReader struct {
	environ []string
}

// NewEnvReader creates a new reader which gets key-value pairs from the environment
func NewEnvReader(opts ...func(*envReader)) envReader {
	r := envReader{}
	for _, opt := range opts {
		opt(&r)
	}

	if r.environ == nil {
		r.environ = os.Environ()
	}

	return r
}

func (r envReader) Read() (ReadResult, error) {
	envAsMap := make(map[string]string, len(r.environ))
	for _, entry := range r.environ {
		if entry == "" {
			continue
		}

		key, value, _ := split(entry)
		envAsMap[key] = value
	}
	return NewSimpleReadResult(envAsMap), nil
}

func split(entry string) (string, string, error) {
	parts := strings.SplitN(entry, "=", 2)
	switch len(parts) {
	case 0:
		return "", "", fmt.Errorf("cannot split empty string")
	case 1:
		return parts[0], "", nil
	default:
		return parts[0], parts[1], nil
	}
}

func WithEnviron(environ []string) func(*envReader) {
	return func(r *envReader) {
		r.environ = environ
	}
}
