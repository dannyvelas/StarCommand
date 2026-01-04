/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"fmt"
	"os"

	"github.com/dannyvelas/homelab/cmd"
	"github.com/dannyvelas/homelab/env"
	"github.com/dannyvelas/homelab/helpers"
)

func main() {
	envVars, missingVars := env.New()
	if len(missingVars) > 0 {
		fmt.Fprintf(os.Stderr, "Required environment variables are missing:%s\n", helpers.StringSliceToBulletedList(missingVars))
		os.Exit(1)
	}

	cmd.Initialize(envVars)
	cmd.Execute()
}
