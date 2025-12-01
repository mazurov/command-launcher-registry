package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
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

// StorageConfig holds storage configuration
type StorageConfig struct {
	Type string `mapstructure:"type"` // file | oci | s3 (future)
	Path string `mapstructure:"path"` // for file storage
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

// Load loads configuration from environment variables, config file, and defaults
func Load(configFile string) (*Config, error) {
	v := viper.New()

	// Set defaults
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("storage.type", "file")
	v.SetDefault("storage.path", "./data/registry.json")
	v.SetDefault("auth.type", "none")
	v.SetDefault("auth.users_file", "./data/users.yaml")
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")

	// Bind environment variables with COLA_REGISTRY_ prefix
	v.SetEnvPrefix("COLA_REGISTRY")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Load config file if provided
	if configFile != "" {
		v.SetConfigFile(configFile)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Unmarshal into config struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate server config
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("server.port must be between 1 and 65535")
	}

	// Validate storage type
	if c.Storage.Type != "file" {
		return fmt.Errorf("storage.type must be 'file' (oci and s3 not yet supported)")
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
