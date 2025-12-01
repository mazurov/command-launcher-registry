package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/criteo/command-launcher-registry/internal/auth"
	"github.com/criteo/command-launcher-registry/internal/config"
	"github.com/criteo/command-launcher-registry/internal/server"
	"github.com/criteo/command-launcher-registry/internal/server/handlers"
	"github.com/criteo/command-launcher-registry/internal/storage"
)

var configFile string

// ServerCmd represents the server command
var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the COLA registry HTTP server",
	Long:  `Start the HTTP server that serves Command Launcher registry index and provides REST API for registry management.`,
	RunE:  runServer,
}

func init() {
	ServerCmd.Flags().StringVarP(&configFile, "config", "c", "", "Path to configuration file (optional, can also use COLA_REGISTRY_CONFIG_FILE env var)")
}

func runServer(cmd *cobra.Command, args []string) error {
	// Check for config file from environment variable if not provided via flag
	if configFile == "" {
		configFile = os.Getenv("COLA_REGISTRY_CONFIG_FILE")
	}

	// Load configuration
	cfg, err := config.Load(configFile)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Create logger
	logger := server.NewLogger(cfg.Logging.Level, cfg.Logging.Format)

	// Log startup
	logger.Info("Server starting",
		"version", "1.0.0",
		"port", cfg.Server.Port,
		"config_file", configFile,
		"storage_type", cfg.Storage.Type,
		"storage_path", cfg.Storage.Path,
		"auth_type", cfg.Auth.Type)

	// Initialize storage
	store, err := storage.NewFileStorage(cfg.Storage.Path, logger)
	if err != nil {
		logger.Error("Failed to initialize storage",
			"error", err,
			"storage_path", cfg.Storage.Path)
		return fmt.Errorf("failed to initialize storage: %w", err)
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
			return fmt.Errorf("failed to initialize basic auth: %w", err)
		}
	default:
		return fmt.Errorf("unsupported auth type: %s", cfg.Auth.Type)
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
		return err
	}

	return nil
}
