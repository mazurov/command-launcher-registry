package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/criteo/command-launcher-registry/internal/client/auth"
)

const (
	// URLEnvVar is the environment variable for server URL
	URLEnvVar = "COLA_REGISTRY_URL"
)

// ResolveURL resolves the server URL using precedence:
// 1. flagURL (--url flag)
// 2. Environment variable (COLA_REGISTRY_URL)
// 3. Stored URL from credentials file
// Returns error if no URL found
func ResolveURL(flagURL string) (string, error) {
	// Priority 1: CLI flag
	if flagURL != "" {
		return NormalizeURL(flagURL), nil
	}

	// Priority 2: Environment variable
	if envURL := os.Getenv(URLEnvVar); envURL != "" {
		return NormalizeURL(envURL), nil
	}

	// Priority 3: Stored URL
	storedURL, err := auth.LoadStoredURL()
	if err != nil {
		return "", fmt.Errorf("no server URL configured. Use --url flag, %s env var, or run 'login' command", URLEnvVar)
	}

	return NormalizeURL(storedURL), nil
}

// NormalizeURL removes trailing slashes from URLs
func NormalizeURL(url string) string {
	return strings.TrimRight(url, "/")
}
