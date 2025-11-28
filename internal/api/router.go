package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/mazurov/command-launcher-registry/internal/api/handlers"
	"github.com/mazurov/command-launcher-registry/internal/api/middleware"
	"github.com/mazurov/command-launcher-registry/internal/auth"
	"github.com/mazurov/command-launcher-registry/internal/auth/provider"
	githubprovider "github.com/mazurov/command-launcher-registry/internal/auth/provider/github"
	"github.com/mazurov/command-launcher-registry/internal/service"
)

// SetupRouter configures all routes and middleware
func SetupRouter(svc *service.Service, authConfig *auth.Config) *gin.Engine {
	router := gin.Default()

	// Apply global middleware
	router.Use(middleware.CORS())
	router.Use(middleware.RequestLogger())
	router.Use(middleware.ErrorHandler())

	// Initialize auth provider based on strategy
	var authProvider provider.AuthProvider
	if authConfig.Strategy == "github" && authConfig.GitHub != nil {
		authProvider = githubprovider.NewGitHubProvider(authConfig.GitHub)
	} else {
		// Default to no-op or error - for now we require GitHub
		panic(fmt.Sprintf("unsupported auth strategy: %s", authConfig.Strategy))
	}

	// Initialize handlers
	registryHandler := handlers.NewRegistryHandler(svc)
	packageHandler := handlers.NewPackageHandler(svc)
	versionHandler := handlers.NewVersionHandler(svc)
	authHandler := handlers.NewAuthHandler(authProvider, authConfig)

	// Initialize device flow handler and connect to auth handler
	deviceFlowHandler := handlers.NewDeviceFlowHandler(authHandler)
	authHandler.DeviceFlowHandler = deviceFlowHandler

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Root route - API info
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"name":    "Remote Registry API",
			"version": "1.0.0",
			"docs":    "/docs",
		})
	})

	// Auth routes (public - no authentication required)
	auth := router.Group("/auth")
	{
		if authConfig.Strategy == "github" {
			auth.GET("/github/login", authHandler.HandleLogin)
			auth.GET("/github/callback", authHandler.HandleCallback)
			// PAT exchange for CI/CD (non-interactive)
			auth.POST("/github/token", authHandler.ExchangeGitHubPAT)
		}

		// Device flow endpoints (for CLI without local server)
		auth.POST("/device/code", deviceFlowHandler.HandleDeviceCode)
		auth.GET("/device", deviceFlowHandler.HandleDeviceAuthorize)
		auth.POST("/device/token", deviceFlowHandler.HandleDeviceToken)

		// Protected auth routes (require authentication)
		auth.GET("/me", middleware.AuthMiddleware(authConfig.JWTSecret), authHandler.GetCurrentUser)
	}

	// Remote API routes
	v1 := router.Group("/remote")
	{
		// Registry routes
		registries := v1.Group("/registries")
		{
			// Read endpoints - require authentication only
			registries.GET("", middleware.AuthMiddleware(authConfig.JWTSecret), registryHandler.ListRegistries)

			// Write endpoints - require authentication + team membership
			writeTeams := authConfig.GitHub.WriteTeams
			registries.POST("", middleware.AuthMiddleware(authConfig.JWTSecret), middleware.RequireTeamMembership(writeTeams), registryHandler.CreateRegistry)

			// Nested routes under specific registry
			registry := registries.Group("/:name")
			{
				// Read endpoints
				registry.GET("", middleware.AuthMiddleware(authConfig.JWTSecret), registryHandler.GetRegistry)
				// CDT-compatible index endpoint
				registry.GET("/index.json", middleware.AuthMiddleware(authConfig.JWTSecret), registryHandler.GetRegistryIndex)

				// Write endpoints
				registry.PUT("", middleware.AuthMiddleware(authConfig.JWTSecret), middleware.RequireTeamMembership(writeTeams), registryHandler.UpdateRegistry)
				registry.DELETE("", middleware.AuthMiddleware(authConfig.JWTSecret), middleware.RequireTeamMembership(writeTeams), registryHandler.DeleteRegistry)

				// Package routes (nested under registry)
				packages := registry.Group("/packages")
				{
					// Read endpoints
					packages.GET("", middleware.AuthMiddleware(authConfig.JWTSecret), packageHandler.ListPackages)
					packages.GET("/:package", middleware.AuthMiddleware(authConfig.JWTSecret), packageHandler.GetPackage)

					// Write endpoints
					packages.POST("", middleware.AuthMiddleware(authConfig.JWTSecret), middleware.RequireTeamMembership(writeTeams), packageHandler.CreatePackage)
					packages.PUT("/:package", middleware.AuthMiddleware(authConfig.JWTSecret), middleware.RequireTeamMembership(writeTeams), packageHandler.UpdatePackage)
					packages.DELETE("/:package", middleware.AuthMiddleware(authConfig.JWTSecret), middleware.RequireTeamMembership(writeTeams), packageHandler.DeletePackage)

					// Version routes (nested under package)
					versions := packages.Group("/:package/versions")
					{
						// Read endpoints
						versions.GET("", middleware.AuthMiddleware(authConfig.JWTSecret), versionHandler.ListVersions)
						versions.GET("/:version", middleware.AuthMiddleware(authConfig.JWTSecret), versionHandler.GetVersion)

						// Write endpoints
						versions.POST("", middleware.AuthMiddleware(authConfig.JWTSecret), middleware.RequireTeamMembership(writeTeams), versionHandler.PublishVersion)
						versions.DELETE("/:version", middleware.AuthMiddleware(authConfig.JWTSecret), middleware.RequireTeamMembership(writeTeams), versionHandler.DeleteVersion)
					}
				}
			}
		}
	}

	return router
}
