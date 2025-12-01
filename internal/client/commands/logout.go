package commands

import (
	"github.com/criteo/command-launcher-registry/internal/client/auth"
	"github.com/criteo/command-launcher-registry/internal/client/errors"
	"github.com/criteo/command-launcher-registry/internal/client/output"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored credentials",
	Long: `Remove stored credentials for the current server.

This operation is idempotent - it succeeds even if no credentials are stored.`,
	Args: cobra.NoArgs,
	Run:  runLogout,
}

func runLogout(cmd *cobra.Command, args []string) {
	// Delete credentials (idempotent - succeeds even if no credentials exist)
	if err := auth.DeleteCredentials(); err != nil {
		errors.ExitWithError(err, "failed to remove credentials")
	}

	if flagJSON {
		output.OutputJSON(map[string]bool{"logged_out": true}, nil)
	} else {
		output.PrintSuccess("Logged out successfully")
	}
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
