package cli

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/criteo/command-launcher-registry/internal/auth"
	"github.com/criteo/command-launcher-registry/internal/config"
	"github.com/criteo/command-launcher-registry/internal/server"
	"github.com/criteo/command-launcher-registry/internal/server/handlers"
	"github.com/criteo/command-launcher-registry/internal/storage"
)

// Exit codes
const (
	ExitCodeOK                   = 0
	ExitCodeInvalidConfig        = 1
	ExitCodeStorageInitFailed    = 2
	ExitCodeServerStartupFailed  = 3
)

var v *viper.Viper

// ServerCmd represents the server command
var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the COLA registry HTTP server",
	Long:  `Start the HTTP server that serves Command Launcher registry index and provides REST API for registry management.`,
	RunE:  runServer,
}

func init() {
	v = config.NewViper()

	// CLI flags - these take precedence over environment variables
	ServerCmd.Flags().String("storage-uri", "", "Storage URI (e.g., file://./data/registry.json)")
	ServerCmd.Flags().String("storage-token", "", "Storage authentication token (passed to storage backend)")
	ServerCmd.Flags().Int("port", 0, "Server port")
	ServerCmd.Flags().String("host", "", "Bind address")
	ServerCmd.Flags().String("log-level", "", "Log level (debug|info|warn|error)")
	ServerCmd.Flags().String("log-format", "", "Log format (json|text)")
	ServerCmd.Flags().String("auth-type", "", "Authentication type (none|basic)")

	// Bind CLI flags to viper
	v.BindPFlag("storage.uri", ServerCmd.Flags().Lookup("storage-uri"))
	v.BindPFlag("storage.token", ServerCmd.Flags().Lookup("storage-token"))
	v.BindPFlag("server.port", ServerCmd.Flags().Lookup("port"))
	v.BindPFlag("server.host", ServerCmd.Flags().Lookup("host"))
	v.BindPFlag("logging.level", ServerCmd.Flags().Lookup("log-level"))
	v.BindPFlag("logging.format", ServerCmd.Flags().Lookup("log-format"))
	v.BindPFlag("auth.type", ServerCmd.Flags().Lookup("auth-type"))
}

func runServer(cmd *cobra.Command, args []string) error {
	// Load configuration (CLI flags > env vars > defaults)
	cfg, err := config.LoadWithViper(v)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load configuration: %v\n", err)
		os.Exit(ExitCodeInvalidConfig)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid configuration: %v\n", err)
		os.Exit(ExitCodeInvalidConfig)
	}

	// Create logger
	logger := server.NewLogger(cfg.Logging.Level, cfg.Logging.Format)

	// Log effective configuration at startup (with masked token)
	logEffectiveConfig(cfg, logger)

	// Parse storage URI
	storageURI, err := cfg.GetParsedStorageURI()
	if err != nil {
		logger.Error("Failed to parse storage URI",
			"error", err,
			"storage_uri", cfg.Storage.URI)
		os.Exit(ExitCodeInvalidConfig)
	}

	// Initialize storage using factory
	store, err := storage.NewStorage(storageURI, cfg.Storage.Token, logger)
	if err != nil {
		logger.Error("Failed to initialize storage",
			"error", err,
			"storage_uri", cfg.Storage.URI,
			"scheme", storageURI.Scheme)
		os.Exit(ExitCodeStorageInitFailed)
	}

	// Initialize authenticator
	var authenticator auth.Authenticator
	switch cfg.Auth.Type {
	case "none":
		authenticator = auth.NewNoAuth()
		logger.Info("Authentication disabled (auth.type=none)")
	case "basic":
		authenticator, err = auth.NewBasicAuth(cfg.Auth.UsersFile, logger)
		if err != nil {
			logger.Error("Failed to initialize basic auth",
				"error", err,
				"users_file", cfg.Auth.UsersFile)
			os.Exit(ExitCodeStorageInitFailed)
		}
	default:
		logger.Error("Unsupported auth type", "auth_type", cfg.Auth.Type)
		os.Exit(ExitCodeInvalidConfig)
	}

	// Create server
	srv := server.NewServer(cfg, logger, store, authenticator)

	// Create all handlers
	indexHandler := handlers.NewIndexHandler(store, logger)
	registryHandler := handlers.NewRegistryHandler(store, logger)
	packageHandler := handlers.NewPackageHandler(store, logger)
	versionHandler := handlers.NewVersionHandler(store, logger)
	healthHandler := handlers.NewHealthHandler(store, logger)
	metricsHandler := handlers.NewMetricsHandler(logger)
	whoamiHandler := handlers.NewWhoamiHandler(authenticator, logger)

	// Set all handlers
	srv.SetHandlers(server.HandlerSet{
		IndexGet:       indexHandler.GetIndex,
		IndexOptions:   indexHandler.HandleOptions,
		Health:         healthHandler.GetHealth,
		Metrics:        metricsHandler.GetMetrics,
		Whoami:         whoamiHandler.GetWhoami,
		ListRegistries: registryHandler.ListRegistries,
		CreateRegistry: registryHandler.CreateRegistry,
		GetRegistry:    registryHandler.GetRegistry,
		UpdateRegistry: registryHandler.UpdateRegistry,
		DeleteRegistry: registryHandler.DeleteRegistry,
		ListPackages:   packageHandler.ListPackages,
		CreatePackage:  packageHandler.CreatePackage,
		GetPackage:     packageHandler.GetPackage,
		UpdatePackage:  packageHandler.UpdatePackage,
		DeletePackage:  packageHandler.DeletePackage,
		ListVersions:   versionHandler.ListVersions,
		CreateVersion:  versionHandler.CreateVersion,
		GetVersion:     versionHandler.GetVersion,
		DeleteVersion:  versionHandler.DeleteVersion,
	})

	// Start server
	logger.Info("Server ready to accept connections",
		"address", fmt.Sprintf("http://%s:%d", cfg.Server.Host, cfg.Server.Port))

	if err := srv.Start(); err != nil {
		logger.Error("Server stopped with error", "error", err)
		os.Exit(ExitCodeServerStartupFailed)
	}

	return nil
}

// logEffectiveConfig logs the effective configuration at startup
func logEffectiveConfig(cfg *config.Config, logger *slog.Logger) {
	tokenDisplay := cfg.MaskToken()
	if tokenDisplay == "" {
		tokenDisplay = "(not set)"
	}

	logger.Info("Server starting with configuration",
		"version", "1.0.0",
		"storage_uri", cfg.Storage.URI,
		"storage_token", tokenDisplay,
		"port", cfg.Server.Port,
		"host", cfg.Server.Host,
		"log_level", cfg.Logging.Level,
		"log_format", cfg.Logging.Format,
		"auth_type", cfg.Auth.Type,
		"auth_users_file", cfg.Auth.UsersFile,
	)
}
