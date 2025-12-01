package commands

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/criteo/command-launcher-registry/internal/client"
	"github.com/criteo/command-launcher-registry/internal/client/auth"
	"github.com/criteo/command-launcher-registry/internal/client/config"
	"github.com/criteo/command-launcher-registry/internal/client/errors"
	"github.com/criteo/command-launcher-registry/internal/client/output"
	"github.com/spf13/cobra"
)

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show authentication status and server information",
	Long: `Check authentication status by calling the server's /api/v1/whoami endpoint.

Resolves server URL and credentials using normal precedence:
- URL: --url flag > COLA_REGISTRY_URL env var > stored URL
- Token: --token flag > COLA_REGISTRY_SESSION_TOKEN env var > stored token`,
	Args: cobra.NoArgs,
	Run:  runWhoami,
}

func runWhoami(cmd *cobra.Command, args []string) {
	// Resolve URL
	serverURL, err := config.ResolveURL(flagURL)
	if err != nil {
		errors.ExitWithCode(errors.ExitInvalidArguments, err.Error())
	}

	// Resolve token
	token, err := auth.ResolveToken(flagToken)
	if err != nil {
		errors.ExitWithError(err, "failed to resolve authentication token")
	}

	// Check authentication by calling /api/v1/whoami
	encodedToken := ""
	if token != "" {
		encodedToken = base64.StdEncoding.EncodeToString([]byte(token))
	}

	c := client.NewClient(serverURL, encodedToken, flagTimeout, flagVerbose)
	resp, err := c.Get("/api/v1/whoami")
	if err != nil {
		errors.ExitWithError(err, "failed to connect to server")
	}
	defer resp.Body.Close()

	authenticated := resp.StatusCode == http.StatusOK
	var username string

	// Extract username from server response
	if resp.StatusCode == http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		var whoamiResp map[string]interface{}
		if json.Unmarshal(body, &whoamiResp) == nil {
			if user, ok := whoamiResp["username"].(string); ok {
				username = user
			}
		}
		// If username is still empty, set a fallback message
		if username == "" {
			username = "(username unknown)"
		}
	}

	if flagJSON {
		output.OutputJSON(map[string]interface{}{
			"server":        serverURL,
			"authenticated": authenticated,
			"username":      username,
		}, nil)
	} else {
		if authenticated {
			output.PrintSuccess(fmt.Sprintf("Authenticated to %s as %s", serverURL, username))
		} else if resp.StatusCode == http.StatusUnauthorized {
			output.PrintError(fmt.Sprintf("Not authenticated to %s", serverURL))
			fmt.Println("Run 'cola-regctl login' to authenticate")
		} else {
			output.PrintError(fmt.Sprintf("Server returned status %d", resp.StatusCode))
		}
	}

	if !authenticated {
		errors.ExitWithCode(errors.ExitAuthError, "")
	}
}

func init() {
	rootCmd.AddCommand(whoamiCmd)
}
