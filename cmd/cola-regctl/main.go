package main

import (
	"os"

	"github.com/criteo/command-launcher-registry/internal/client/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		os.Exit(1)
	}
}
