package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/dannyvelas/conflux"
	"github.com/dannyvelas/starcommand/internal/helpers"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "error loading .env file: %v", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	configMux := conflux.NewConfigMux(
		conflux.WithYAMLFileReader(helpers.ConfigFile),
		conflux.WithEnvReader(),
		conflux.WithBitwardenSecretReader(),
	)

	rootCmd := rootCmd(configMux)
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
