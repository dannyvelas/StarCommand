package config

import (
	"fmt"
	"os"
	"strings"
)

var _ unvalidatedReader = envReader{}

type envReader struct{}

func newEnvReader() envReader {
	return envReader{}
}

func (p envReader) ReadUnvalidated() (map[string]string, error) {
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
