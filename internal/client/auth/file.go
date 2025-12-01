//go:build linux
// +build linux

package auth

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var ErrNotFound = errors.New("credentials not found")

const (
	configDir  = ".config/cola-registry"
	configFile = "credentials.yaml"
)

// Credentials represents the stored credentials
type Credentials struct {
	URL   string `yaml:"url"`
	Token string `yaml:"token"`
}

// getConfigPath returns the path to the credentials file
func getConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, configDir, configFile), nil
}

// LoadCredentials loads credentials from the file
func LoadCredentials() (*Credentials, error) {
	path, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to read credentials file: %w", err)
	}

	var creds Credentials
	if err := yaml.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("failed to parse credentials file: %w", err)
	}

	return &creds, nil
}

// SaveCredentials saves credentials to the file
func SaveCredentials(url, token string) error {
	path, err := getConfigPath()
	if err != nil {
		return err
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	creds := Credentials{
		URL:   url,
		Token: token,
	}

	data, err := yaml.Marshal(&creds)
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	// Write with 0600 permissions (read/write for owner only)
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write credentials file: %w", err)
	}

	return nil
}

// DeleteCredentials removes the credentials file
func DeleteCredentials() error {
	path, err := getConfigPath()
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete credentials file: %w", err)
	}

	return nil
}

// LoadStoredToken loads just the token from credentials
func LoadStoredToken() (string, error) {
	creds, err := LoadCredentials()
	if err != nil {
		return "", err
	}
	if creds.Token == "" {
		return "", ErrNotFound
	}
	return creds.Token, nil
}

// LoadStoredURL loads just the URL from credentials
func LoadStoredURL() (string, error) {
	creds, err := LoadCredentials()
	if err != nil {
		return "", err
	}
	if creds.URL == "" {
		return "", ErrNotFound
	}
	return creds.URL, nil
}
