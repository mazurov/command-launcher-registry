package config

import (
	"fmt"

	"github.com/mazurov/command-launcher-registry/internal/auth"
	"github.com/spf13/viper"
)

// Config holds application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     auth.Config
	Log      LogConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port string
	Mode string // debug, release
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Type string // postgres, sqlite
	DSN  string // connection string
}

// AuthConfig holds authentication configuration (deprecated, kept for reference)
// Use auth.Config instead
type AuthConfig struct {
	JWTSecret string
	APIKeys   []string
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level  string // debug, info, warn, error
	Format string // json, text
}

// LoadConfig loads configuration from viper (which already has flags/env bound)
func LoadConfig() (*Config, error) {
	// Use the global viper instance that has flags already bound from main.go
	config := &Config{
		Server: ServerConfig{
			Port: viper.GetString("port"),
			Mode: viper.GetString("mode"),
		},
		Database: DatabaseConfig{
			Type: viper.GetString("db-type"),
			DSN:  viper.GetString("db-dsn"),
		},
		Auth: auth.Config{
			Strategy:    viper.GetString("auth-strategy"),
			JWTSecret:   viper.GetString("jwt-secret"),
			TokenExpiry: viper.GetInt("token-expiry"),
			GitHub: &auth.GitHubConfig{
				Organization: viper.GetString("github-org"),
				ClientID:     viper.GetString("github-client-id"),
				ClientSecret: viper.GetString("github-client-secret"),
				RedirectURL:  viper.GetString("github-redirect-url"),
				WriteTeams:   viper.GetStringSlice("github-write-teams"),
				Scopes:       viper.GetStringSlice("github-scopes"),
			},
		},
		Log: LogConfig{
			Level:  viper.GetString("log-level"),
			Format: viper.GetString("log-format"),
		},
	}

	// Validate configuration
	if err := validateConfig(config); err != nil {
		return nil, err
	}

	return config, nil
}

func validateConfig(cfg *Config) error {
	if cfg.Database.Type != "postgres" && cfg.Database.Type != "sqlite" {
		return fmt.Errorf("unsupported database type: %s (must be postgres or sqlite)", cfg.Database.Type)
	}

	if cfg.Database.DSN == "" {
		return fmt.Errorf("database DSN is required")
	}

	return nil
}
