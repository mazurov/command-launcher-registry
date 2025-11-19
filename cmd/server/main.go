package main

import (
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
	rootCmd.PersistentFlags().String("jwt-secret", "", "JWT signing secret")
	rootCmd.PersistentFlags().String("log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().String("log-format", "text", "Log format (text, json)")
	rootCmd.PersistentFlags().String("mode", "debug", "Server mode (debug, release)")

	// Bind flags to viper
	_ = viper.BindPFlags(rootCmd.PersistentFlags())
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

	// Setup router
	router := api.SetupRouter(svc, cfg.Auth.JWTSecret)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Starting server on %s", addr)
	log.Printf("Remote API available at: http://localhost%s/remote", addr)
	log.Printf("Health check at: http://localhost%s/health", addr)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
