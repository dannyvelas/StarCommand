package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/dannyvelas/starcommand/internal/models"
	"github.com/goccy/go-yaml"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "error loading .env file: %v\n", err)
		os.Exit(1)
	}

	c, err := newConfig("stc.yml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err.Error())
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	rootCmd := rootCmd(c)
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}

func newConfig(file string) (*models.Config, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("error reading config file %q: %w", file, err)
	}

	var c models.Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("error parsing config file %q: %w", file, err)
	}

	return &c, nil
}
