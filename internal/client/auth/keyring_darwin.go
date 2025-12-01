//go:build darwin
// +build darwin

package auth

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/zalando/go-keyring"
	"gopkg.in/yaml.v3"
)

var ErrNotFound = errors.New("credentials not found")

const (
	keychainService = "cola-registry"
	configDir       = ".config/cola-registry"
	configFile      = "credentials.yaml"
)

// ConfigFile represents the URL-only config file on macOS/Windows
type ConfigFile struct {
	URL string `yaml:"url"`
}

// getConfigPath returns the path to the config file (URL only)
func getConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, configDir, configFile), nil
}

// LoadStoredToken loads the token from macOS Keychain
func LoadStoredToken() (string, error) {
	// Get URL to use as keychain account
	url, err := LoadStoredURL()
	if err != nil {
		return "", err
	}

	// Get token from keychain
	token, err := keyring.Get(keychainService, url)
	if err != nil {
		if err == keyring.ErrNotFound {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("failed to get token from keychain: %w", err)
	}

	return token, nil
}

// LoadStoredURL loads the URL from config file
func LoadStoredURL() (string, error) {
	path, err := getConfigPath()
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("failed to read config file: %w", err)
	}

	var config ConfigFile
	if err := yaml.Unmarshal(data, &config); err != nil {
		return "", fmt.Errorf("failed to parse config file: %w", err)
	}

	if config.URL == "" {
		return "", ErrNotFound
	}

	return config.URL, nil
}

// SaveCredentials saves URL to config file and token to Keychain
func SaveCredentials(url, token string) error {
	// Save URL to config file
	path, err := getConfigPath()
	if err != nil {
		return err
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	config := ConfigFile{URL: url}
	data, err := yaml.Marshal(&config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// Save token to keychain
	if err := keyring.Set(keychainService, url, token); err != nil {
		return fmt.Errorf("failed to save token to keychain: %w", err)
	}

	return nil
}

// DeleteCredentials removes URL from config and token from Keychain
func DeleteCredentials() error {
	// Get URL first (needed to delete from keychain)
	url, urlErr := LoadStoredURL()

	// Delete token from keychain if URL was found
	if urlErr == nil {
		if err := keyring.Delete(keychainService, url); err != nil && err != keyring.ErrNotFound {
			return fmt.Errorf("failed to delete token from keychain: %w", err)
		}
	}

	// Delete config file
	path, err := getConfigPath()
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete config file: %w", err)
	}

	return nil
}
