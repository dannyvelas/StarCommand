package config

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

func NewConfig(file string) (*Config, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("error reading config file %q: %w", file, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("error parsing config file %q: %w", file, err)
	}

	return &cfg, nil
}
