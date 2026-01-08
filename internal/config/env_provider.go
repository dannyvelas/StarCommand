package config

import (
	"fmt"
	"os"
	"strings"
)

var (
	_ provider          = envProvider{}
	_ unvalidatedReader = envProvider{}
)

type envProvider struct{}

func newEnvProvider() envProvider {
	return envProvider{}
}

func (p envProvider) UnmarshalInto(target any) error {
	// read env
	envMap, err := p.ReadUnvalidated()
	if err != nil {
		return fmt.Errorf("error reading env: %v", err)
	}

	// decode env
	if err := decode(envMap, target); err != nil {
		return fmt.Errorf("error decoding env into a map: %v", err)
	}
	return nil
}

func (p envProvider) ReadUnvalidated() (map[string]string, error) {
	environ := os.Environ()
	envAsMap := make(map[string]string, len(environ))
	for _, entry := range environ {
		if entry != "" {
			key, value, err := split(entry)
			if err != nil {
				return nil, fmt.Errorf("error splitting: %v", err)
			}
			envAsMap[key] = value
		}
	}
	return envAsMap, nil
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
