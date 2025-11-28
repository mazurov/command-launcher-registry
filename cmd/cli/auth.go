package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication commands",
	Long:  `Manage authentication with the registry`,
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to the registry via GitHub OAuth",
	Long:  `Interactive login via GitHub OAuth. Opens browser automatically and starts local callback server.`,
	RunE:  runLogin,
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout from the registry",
	Long:  `Remove stored authentication token`,
	RunE:  runLogout,
}

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show current authenticated user",
	Long:  `Display information about the currently authenticated user`,
	RunE:  runWhoami,
}

func init() {
	authCmd.AddCommand(loginCmd)
	authCmd.AddCommand(logoutCmd)
	authCmd.AddCommand(whoamiCmd)
	rootCmd.AddCommand(authCmd)
}

// TokenStorage represents the stored token configuration
type TokenStorage struct {
	Token     string    `yaml:"token"`
	ExpiresAt time.Time `yaml:"expires_at"`
	User      UserInfo  `yaml:"user"`
}

// UserInfo represents user information from the token
type UserInfo struct {
	Username string   `yaml:"username"`
	Email    string   `yaml:"email"`
	Teams    []string `yaml:"teams"`
	Provider string   `yaml:"provider"`
}

// DeviceCodeResponse represents the response from /auth/device/code
type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

// DeviceTokenResponse represents the response from /auth/device/token
type DeviceTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	User        struct {
		ID       string   `json:"id"`
		Username string   `json:"username"`
		Email    string   `json:"email"`
		Teams    []string `json:"teams"`
		Provider string   `json:"provider"`
	} `json:"user"`
}

// DeviceTokenError represents error response from /auth/device/token
type DeviceTokenError struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// getConfigPath returns the path to the CLI configuration file
func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "cola-registry")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return filepath.Join(configDir, "config.yaml"), nil
}

// loadToken loads the stored token from config file
func loadToken() (*TokenStorage, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No token stored
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var storage TokenStorage
	if err := yaml.Unmarshal(data, &storage); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Check if token is expired
	if time.Now().After(storage.ExpiresAt) {
		return nil, nil // Token expired
	}

	return &storage, nil
}

// saveToken saves the token to config file
func saveToken(storage *TokenStorage) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(storage)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// deleteToken removes the stored token
func deleteToken() error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete config file: %w", err)
	}

	return nil
}

// openBrowser opens the default browser with the given URL
func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}

// runLogin performs device flow login (no local server needed)
func runLogin(cmd *cobra.Command, args []string) error {
	fmt.Println("Starting device flow login...")

	// Step 1: Request device code from server
	resp, err := http.Post(apiURL+"/auth/device/code", "application/json", nil)
	if err != nil {
		return fmt.Errorf("failed to request device code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned error: %s - %s", resp.Status, string(body))
	}

	var deviceCode DeviceCodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&deviceCode); err != nil {
		return fmt.Errorf("failed to parse device code response: %w", err)
	}

	// Step 2: Show user code and URL
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Printf("  Please visit: %s\n", deviceCode.VerificationURI)
	fmt.Printf("  And enter code: %s\n", deviceCode.UserCode)
	fmt.Println(strings.Repeat("=", 60) + "\n")

	// Step 3: Open browser automatically
	verifyURL := deviceCode.VerificationURI + "?user_code=" + deviceCode.UserCode
	if err := openBrowser(verifyURL); err != nil {
		fmt.Printf("Failed to open browser automatically: %v\n", err)
		fmt.Printf("Please open the URL manually.\n\n")
	} else {
		fmt.Println("Browser opened. Please authorize the application...\n")
	}

	// Step 4: Poll for token
	fmt.Println("Waiting for authorization...")
	interval := time.Duration(deviceCode.Interval) * time.Second
	expiresAt := time.Now().Add(time.Duration(deviceCode.ExpiresIn) * time.Second)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Check if expired
			if time.Now().After(expiresAt) {
				return fmt.Errorf("device code expired - please try again")
			}

			// Poll for token
			token, user, err := pollForToken(deviceCode.DeviceCode)
			if err != nil {
				// Check if it's "authorization_pending" - continue waiting
				if strings.Contains(err.Error(), "authorization_pending") {
					continue
				}
				return err
			}

			// Success! Save token
			storage := &TokenStorage{
				Token:     token,
				ExpiresAt: expiresAt,
				User: UserInfo{
					Username: user.Username,
					Email:    user.Email,
					Teams:    user.Teams,
					Provider: user.Provider,
				},
			}

			if err := saveToken(storage); err != nil {
				return fmt.Errorf("failed to save token: %w", err)
			}

			fmt.Printf("\n✓ Successfully authenticated as %s\n", user.Username)
			if user.Email != "" {
				fmt.Printf("  Email: %s\n", user.Email)
			}
			if len(user.Teams) > 0 {
				fmt.Printf("  Teams: %v\n", user.Teams)
			}
			fmt.Printf("\nToken saved to: %s\n", mustGetConfigPath())

			return nil
		}
	}
}

// pollForToken polls the server for an access token
func pollForToken(deviceCode string) (string, *UserInfo, error) {
	reqBody := map[string]string{"device_code": deviceCode}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(apiURL+"/auth/device/token", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", nil, fmt.Errorf("failed to poll for token: %w", err)
	}
	defer resp.Body.Close()

	// Success case
	if resp.StatusCode == http.StatusOK {
		var tokenResp DeviceTokenResponse
		if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
			return "", nil, fmt.Errorf("failed to parse token response: %w", err)
		}

		user := &UserInfo{
			Username: tokenResp.User.Username,
			Email:    tokenResp.User.Email,
			Teams:    tokenResp.User.Teams,
			Provider: tokenResp.User.Provider,
		}

		return tokenResp.AccessToken, user, nil
	}

	// Error case
	var errResp DeviceTokenError
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		return "", nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	return "", nil, fmt.Errorf("%s: %s", errResp.Error, errResp.Message)
}

// runLogout removes stored authentication
func runLogout(cmd *cobra.Command, args []string) error {
	if err := deleteToken(); err != nil {
		return fmt.Errorf("failed to logout: %w", err)
	}

	fmt.Println("✓ Logged out successfully")
	fmt.Printf("Token removed from: %s\n", mustGetConfigPath())

	return nil
}

// runWhoami shows current authenticated user
func runWhoami(cmd *cobra.Command, args []string) error {
	// Try to get token from flag first
	token := jwtToken
	if token == "" {
		// Try to load from storage
		storage, err := loadToken()
		if err != nil {
			return fmt.Errorf("failed to load token: %w", err)
		}
		if storage == nil {
			return fmt.Errorf("not authenticated - run 'cola-registry-cli auth login' first")
		}
		token = storage.Token
	}

	// Get user info from /auth/me endpoint
	req, err := http.NewRequest("GET", apiURL+"/auth/me", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned error: %s - %s", resp.Status, string(body))
	}

	var userInfo struct {
		ID       string   `json:"id"`
		Username string   `json:"username"`
		Email    string   `json:"email"`
		Teams    []string `json:"teams"`
		Provider string   `json:"provider"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	fmt.Printf("Authenticated as: %s\n", userInfo.Username)
	if userInfo.Email != "" {
		fmt.Printf("Email: %s\n", userInfo.Email)
	}
	fmt.Printf("Provider: %s\n", userInfo.Provider)
	if len(userInfo.Teams) > 0 {
		fmt.Printf("Teams: %v\n", userInfo.Teams)
	} else {
		fmt.Println("Teams: (none)")
	}

	return nil
}

// mustGetConfigPath returns config path or empty string on error
func mustGetConfigPath() string {
	path, _ := getConfigPath()
	return path
}
