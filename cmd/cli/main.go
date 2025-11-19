package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	apiURL string
	apiKey string

	rootCmd = &cobra.Command{
		Use:   "registry-cli",
		Short: "Remote Registry CLI",
		Long:  `Command-line interface for managing the remote registry`,
	}
)

func init() {
	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", "http://localhost:8080", "Registry API URL")
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "API key for authentication")

	// Add subcommands
	rootCmd.AddCommand(registryCmd)
	rootCmd.AddCommand(packageCmd)
	rootCmd.AddCommand(versionCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
