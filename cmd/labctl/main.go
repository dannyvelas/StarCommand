package main

import (
	"fmt"
	"os"

	"github.com/dannyvelas/homelab/internal/env"
	"github.com/dannyvelas/homelab/internal/helpers"
)

func main() {
	envVars, missingVars := env.New()
	if len(missingVars) > 0 {
		fmt.Fprintf(os.Stderr, "Required environment variables are missing:%s\n", helpers.StringSliceToBulletedList(missingVars))
		os.Exit(1)
	}

	initialize(envVars)
	execute()
}
