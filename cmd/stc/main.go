package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/dannyvelas/starcommand/internal/config"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "error loading .env file: %v", err)
		os.Exit(1)
	}

	c, err := config.NewConfig("stc.yml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	rootCmd := rootCmd(c)
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
