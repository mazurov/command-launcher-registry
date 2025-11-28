package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	apiURL   string
	jwtToken string

	rootCmd = &cobra.Command{
		Use:   "registry-cli",
		Short: "Remote Registry CLI",
		Long:  `Command-line interface for managing the remote registry`,
	}
)

func init() {
	// Get default values from environment variables
	defaultAPIURL := os.Getenv("COLA_REGISTRY_URL")
	if defaultAPIURL == "" {
		defaultAPIURL = "http://localhost:8080"
	}

	defaultToken := os.Getenv("COLA_REGISTRY_TOKEN")

	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", defaultAPIURL, "Registry API URL (env: COLA_REGISTRY_URL)")
	rootCmd.PersistentFlags().StringVar(&jwtToken, "token", defaultToken, "JWT token for authentication (env: COLA_REGISTRY_TOKEN)")

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
