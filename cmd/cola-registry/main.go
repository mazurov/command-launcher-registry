package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/criteo/command-launcher-registry/internal/cli"
)

var version = "1.0.0"

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "cola-registry",
	Short: "Command Launcher Registry Server",
	Long: `COLA Registry Server provides a REST API for managing Command Launcher
remote registries. It serves registry indexes and provides full CRUD operations
for registries, packages, and versions.`,
	Version: version,
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(cli.ServerCmd)
	rootCmd.AddCommand(cli.AuthCmd)

	// Set version template
	rootCmd.SetVersionTemplate(`{{.Version}}
`)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
