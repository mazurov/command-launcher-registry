package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mazurov/command-launcher-registry/internal/api"
	"github.com/mazurov/command-launcher-registry/internal/api/middleware"
	"github.com/mazurov/command-launcher-registry/internal/config"
	"github.com/mazurov/command-launcher-registry/internal/repository"
	"github.com/mazurov/command-launcher-registry/internal/service"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "registry-server",
		Short: "Remote Registry Server",
		Long:  `A modern package registry server built with Gin and GORM`,
		Run:   runServer,
	}
)

func init() {
	cobra.OnInitialize(initConfig)

	// Configuration flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: config.yaml)")
	rootCmd.PersistentFlags().String("port", "8080", "Server port")
	rootCmd.PersistentFlags().String("db-type", "sqlite", "Database type (postgres, sqlite)")
	rootCmd.PersistentFlags().String("db-dsn", "registry.db", "Database connection string")
	rootCmd.PersistentFlags().String("log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().String("log-format", "text", "Log format (text, json)")
	rootCmd.PersistentFlags().String("mode", "debug", "Server mode (debug, release)")

	// Authentication flags
	rootCmd.PersistentFlags().String("auth-strategy", "github", "Authentication strategy (only github supported)")
	rootCmd.PersistentFlags().String("jwt-secret", "", "JWT signing secret (auto-generated in dev mode)")
	rootCmd.PersistentFlags().Int("token-expiry", 24, "JWT token expiry in hours")

	// GitHub OAuth flags
	rootCmd.PersistentFlags().String("github-org", "", "GitHub organization name")
	rootCmd.PersistentFlags().String("github-client-id", "", "GitHub OAuth client ID")
	rootCmd.PersistentFlags().String("github-client-secret", "", "GitHub OAuth client secret")
	rootCmd.PersistentFlags().String("github-redirect-url", "http://localhost:8080/auth/github/callback", "GitHub OAuth redirect URL")
	rootCmd.PersistentFlags().StringSlice("github-write-teams", []string{}, "GitHub teams with write access (comma-separated)")
	rootCmd.PersistentFlags().StringSlice("github-scopes", []string{"read:org", "user:email"}, "GitHub OAuth scopes")

	// Bind flags to viper
	_ = viper.BindPFlags(rootCmd.PersistentFlags())

	// Set default values
	viper.SetDefault("auth-strategy", "github")
	viper.SetDefault("token-expiry", 24)
	viper.SetDefault("github-scopes", []string{"read:org", "user:email"})
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
		if err := viper.ReadInConfig(); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading config file: %v\n", err)
		}
	}

	viper.SetEnvPrefix("REGISTRY")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
}

func runServer(cmd *cobra.Command, args []string) {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set Gin mode
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Set log level
	middleware.SetLogLevel(cfg.Log.Level)

	// Initialize database
	log.Printf("Connecting to %s database...", cfg.Database.Type)
	db, err := config.InitDatabase(cfg.Database, cfg.Log.Level)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	log.Println("Database connected and migrations completed")

	// Initialize repository and service layers
	repo := repository.NewRepository(db)
	svc := service.NewService(repo)

	// Validate auth configuration
	if cfg.Auth.JWTSecret == "" {
		// In development mode, generate a random JWT secret
		if cfg.Server.Mode != "release" {
			cfg.Auth.JWTSecret = generateRandomSecret()
			log.Printf("WARNING: Using auto-generated JWT secret (dev mode only). Set REGISTRY_JWT_SECRET for production.")
		} else {
			log.Fatal("JWT secret is required in production (set via --jwt-secret or REGISTRY_JWT_SECRET)")
		}
	}
	if cfg.Auth.Strategy == "github" {
		if cfg.Auth.GitHub.Organization == "" {
			log.Fatal("GitHub organization is required (set via --github-org or REGISTRY_GITHUB_ORG)")
		}
		if cfg.Auth.GitHub.ClientID == "" || cfg.Auth.GitHub.ClientSecret == "" {
			log.Fatal("GitHub OAuth credentials required (set via --github-client-id/secret or REGISTRY_GITHUB_CLIENT_ID/SECRET)")
		}
		if len(cfg.Auth.GitHub.WriteTeams) == 0 {
			log.Println("Warning: No write teams configured, all authenticated users will have read-only access")
		}
	}

	// Setup router
	router := api.SetupRouter(svc, &cfg.Auth)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Starting server on %s", addr)
	log.Printf("Auth strategy: %s", cfg.Auth.Strategy)
	if cfg.Auth.Strategy == "github" {
		log.Printf("GitHub OAuth: org=%s, write-teams=%v", cfg.Auth.GitHub.Organization, cfg.Auth.GitHub.WriteTeams)
		log.Printf("OAuth callback: %s", cfg.Auth.GitHub.RedirectURL)
	}
	log.Printf("Remote API available at: http://localhost%s/remote", addr)
	log.Printf("Auth endpoints: http://localhost%s/auth", addr)
	log.Printf("Health check at: http://localhost%s/health", addr)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// generateRandomSecret generates a random 32-byte secret for JWT signing
func generateRandomSecret() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		log.Fatalf("Failed to generate random secret: %v", err)
	}
	return base64.URLEncoding.EncodeToString(b)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
