package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"

	"github.com/criteo/command-launcher-registry/internal/storage"
)

// Config holds all configuration for the server
type Config struct {
	Server  ServerConfig  `mapstructure:"server"`
	Storage StorageConfig `mapstructure:"storage"`
	Auth    AuthConfig    `mapstructure:"auth"`
	Logging LoggingConfig `mapstructure:"logging"`
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

// StorageConfig holds storage configuration (URI-based)
type StorageConfig struct {
	URI   string `mapstructure:"uri"`   // Storage URI (e.g., file://./data/registry.json)
	Token string `mapstructure:"token"` // Opaque token for storage authentication
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Type      string `mapstructure:"type"`       // none | basic
	UsersFile string `mapstructure:"users_file"` // for basic auth
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `mapstructure:"level"`  // debug | info | warn | error
	Format string `mapstructure:"format"` // json | text
}

// Load loads configuration from environment variables and defaults
// CLI flags take precedence and are bound via viper in the CLI layer
func Load() (*Config, error) {
	v := viper.New()

	// Set defaults
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("storage.uri", "file://./data/registry.json")
	v.SetDefault("storage.token", "")
	v.SetDefault("auth.type", "none")
	v.SetDefault("auth.users_file", "./users.yaml")
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")

	// Bind environment variables with COLA_REGISTRY_ prefix
	v.SetEnvPrefix("COLA_REGISTRY")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Unmarshal into config struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// LoadWithViper loads configuration using a pre-configured viper instance
// This allows CLI flags to be bound before loading
func LoadWithViper(v *viper.Viper) (*Config, error) {
	// Unmarshal into config struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// NewViper creates a new viper instance with defaults and environment binding
func NewViper() *viper.Viper {
	v := viper.New()

	// Set defaults
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("storage.uri", "file://./data/registry.json")
	v.SetDefault("storage.token", "")
	v.SetDefault("auth.type", "none")
	v.SetDefault("auth.users_file", "./users.yaml")
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")

	// Bind environment variables with COLA_REGISTRY_ prefix
	v.SetEnvPrefix("COLA_REGISTRY")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	return v
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate server config
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("server.port must be between 1 and 65535")
	}

	// Validate storage URI
	_, err := storage.ParseStorageURI(c.Storage.URI)
	if err != nil {
		return fmt.Errorf("invalid storage URI: %w", err)
	}

	// Validate auth type
	if c.Auth.Type != "none" && c.Auth.Type != "basic" {
		return fmt.Errorf("auth.type must be 'none' or 'basic'")
	}

	// Validate logging level
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[c.Logging.Level] {
		return fmt.Errorf("logging.level must be debug, info, warn, or error")
	}

	// Validate logging format
	if c.Logging.Format != "json" && c.Logging.Format != "text" {
		return fmt.Errorf("logging.format must be json or text")
	}

	return nil
}

// GetParsedStorageURI returns the parsed storage URI
func (c *Config) GetParsedStorageURI() (*storage.StorageURI, error) {
	return storage.ParseStorageURI(c.Storage.URI)
}

// MaskToken returns a masked version of the storage token for logging
func (c *Config) MaskToken() string {
	if c.Storage.Token == "" {
		return ""
	}
	return "***"
}
