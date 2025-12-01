package cli

import (
	"fmt"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/criteo/command-launcher-registry/internal/auth"
)

// AuthCmd represents the auth command
var AuthCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication utilities",
	Long:  `Utilities for managing authentication credentials.`,
}

// HashPasswordCmd represents the hash-password command
var HashPasswordCmd = &cobra.Command{
	Use:   "hash-password",
	Short: "Generate bcrypt hash for a password",
	Long:  `Generate a bcrypt hash for a password to use in users.yaml file.`,
	RunE:  runHashPassword,
}

func init() {
	AuthCmd.AddCommand(HashPasswordCmd)
}

func runHashPassword(cmd *cobra.Command, args []string) error {
	// Prompt for password
	fmt.Print("Enter password: ")

	// Read password with hidden input
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	fmt.Println() // New line after password input

	password := string(passwordBytes)
	if len(password) == 0 {
		return fmt.Errorf("password cannot be empty")
	}

	// Generate bcrypt hash
	hash, err := auth.HashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Output hash
	fmt.Println("\nBcrypt hash (use this in users.yaml):")
	fmt.Println(hash)

	return nil
}
