package config

import (
	"fmt"
	"strings"
)

var _ Reader = envReader{}

type envReader struct {
	envAsMap map[string]string
}

func NewEnvReader(environ []string) envReader {
	envAsMap := make(map[string]string, len(environ))
	for _, entry := range environ {
		if entry == "" {
			continue
		}

		key, value, _ := split(entry)
		envAsMap[key] = value
	}

	return envReader{
		envAsMap: envAsMap,
	}
}

func (r envReader) read() (readResult, error) {
	return simpleReadResult{configMap: r.envAsMap}, nil
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
