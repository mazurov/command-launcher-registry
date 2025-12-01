package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var (
	// Global flags
	flagURL     string
	flagToken   string
	flagJSON    bool
	flagVerbose bool
	flagTimeout time.Duration
	flagYes     bool
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "cola-regctl",
	Short: "Command Launcher Registry CLI Client",
	Long: `cola-regctl is a command-line client for managing Command Launcher remote registries.

It provides full CRUD operations for registries, packages, and versions via the REST API.`,
}

// Execute executes the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags available to all commands
	rootCmd.PersistentFlags().StringVar(&flagURL, "url", "", "Server URL (or use COLA_REGISTRY_URL env var)")
	rootCmd.PersistentFlags().StringVar(&flagToken, "token", "", "Authentication token in 'user:password' format (or use COLA_REGISTRY_SESSION_TOKEN env var)")
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "Output in JSON format")
	rootCmd.PersistentFlags().BoolVar(&flagVerbose, "verbose", false, "Enable verbose logging")
	rootCmd.PersistentFlags().DurationVar(&flagTimeout, "timeout", 30*time.Second, "HTTP request timeout")
	rootCmd.PersistentFlags().BoolVarP(&flagYes, "yes", "y", false, "Skip confirmation prompts")

	// Add subcommands
	// These will be implemented in subsequent tasks
	// rootCmd.AddCommand(loginCmd)
	// rootCmd.AddCommand(logoutCmd)
	// rootCmd.AddCommand(whoamiCmd)
	// rootCmd.AddCommand(registryCmd)
	// rootCmd.AddCommand(packageCmd)
	// rootCmd.AddCommand(versionCmd)
	// rootCmd.AddCommand(completionCmd)
}

// getGlobalFlags returns the global flag values
func getGlobalFlags() (url, token string, jsonOutput, verbose bool, timeout time.Duration, yes bool) {
	return flagURL, flagToken, flagJSON, flagVerbose, flagTimeout, flagYes
}

// printVersion prints version information (placeholder for now)
func printVersion() {
	fmt.Println("cola-regctl version 0.1.0")
	os.Exit(0)
}
