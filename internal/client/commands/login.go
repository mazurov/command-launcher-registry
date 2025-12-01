package commands

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/criteo/command-launcher-registry/internal/client"
	"github.com/criteo/command-launcher-registry/internal/client/auth"
	"github.com/criteo/command-launcher-registry/internal/client/config"
	"github.com/criteo/command-launcher-registry/internal/client/errors"
	"github.com/criteo/command-launcher-registry/internal/client/output"
	"github.com/criteo/command-launcher-registry/internal/client/prompts"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login [server-url]",
	Short: "Authenticate with a registry server",
	Long: `Authenticate with a registry server and store credentials securely.

Server URL can be provided as an argument or via COLA_REGISTRY_URL environment variable.
If both are provided, the argument takes precedence.

Credentials are stored:
- macOS: Token in Keychain, URL in config file
- Windows: Token in Credential Manager, URL in config file
- Linux: Both in config file with 0600 permissions

Only one server's credentials are stored at a time. Logging into a new server
replaces existing credentials.`,
	Args: cobra.MaximumNArgs(1),
	Run:  runLogin,
}

func runLogin(cmd *cobra.Command, args []string) {
	var serverURL string

	// Resolve server URL: argument takes precedence over environment variable
	if len(args) > 0 {
		serverURL = args[0]
	} else {
		// Try to get URL from environment variable
		var err error
		serverURL, err = config.ResolveURL("")
		if err != nil {
			errors.ExitWithCode(errors.ExitInvalidArguments, "no server URL specified. Provide server URL as argument or set COLA_REGISTRY_URL environment variable")
		}
	}

	// Normalize URL (remove trailing slash)
	serverURL = config.NormalizeURL(serverURL)

	// Prompt for credentials
	username, err := prompts.PromptUsername()
	if err != nil {
		errors.ExitWithError(err, "failed to read username")
	}

	password, err := prompts.PromptPassword()
	if err != nil {
		errors.ExitWithError(err, "failed to read password")
	}

	// Format token as "username:password"
	token := fmt.Sprintf("%s:%s", username, password)

	// Test authentication by calling /api/v1/whoami
	c := client.NewClient(serverURL, base64.StdEncoding.EncodeToString([]byte(token)), flagTimeout, flagVerbose)
	resp, err := c.Get("/api/v1/whoami")
	if err != nil {
		errors.ExitWithError(err, "failed to connect to server")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		errors.ExitWithCode(errors.ExitAuthError, "authentication failed: invalid credentials")
	}

	if resp.StatusCode != http.StatusOK {
		errors.HandleHTTPError(resp.StatusCode, fmt.Sprintf("server returned status %d", resp.StatusCode))
	}

	// Authentication successful - store credentials
	if err := auth.SaveCredentials(serverURL, token); err != nil {
		errors.ExitWithError(err, "failed to save credentials")
	}

	if flagJSON {
		output.OutputJSON(map[string]string{
			"server": serverURL,
			"user":   username,
		}, nil)
	} else {
		output.PrintSuccess(fmt.Sprintf("Logged in to %s as %s", serverURL, username))
	}
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
