package api

import (
	"github.com/gin-gonic/gin"
	"github.com/mazurov/command-launcher-registry/internal/api/handlers"
	"github.com/mazurov/command-launcher-registry/internal/api/middleware"
	"github.com/mazurov/command-launcher-registry/internal/service"
)

// SetupRouter configures all routes and middleware
func SetupRouter(svc *service.Service, jwtSecret string) *gin.Engine {
	router := gin.Default()

	// Apply global middleware
	router.Use(middleware.CORS())
	router.Use(middleware.RequestLogger())
	router.Use(middleware.ErrorHandler())

	// Initialize handlers
	registryHandler := handlers.NewRegistryHandler(svc)
	packageHandler := handlers.NewPackageHandler(svc)
	versionHandler := handlers.NewVersionHandler(svc)

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

	// Remote API routes
	v1 := router.Group("/remote")
	{
		// Registry routes
		registries := v1.Group("/registries")
		{
			registries.POST("", registryHandler.CreateRegistry)
			registries.GET("", registryHandler.ListRegistries)

			// Nested routes under specific registry
			registry := registries.Group("/:name")
			{
				registry.GET("", registryHandler.GetRegistry)
				registry.PUT("", registryHandler.UpdateRegistry)
				registry.DELETE("", registryHandler.DeleteRegistry)

				// CDT-compatible index endpoint
				registry.GET("/index.json", registryHandler.GetRegistryIndex)

				// Package routes (nested under registry)
				packages := registry.Group("/packages")
				{
					packages.POST("", packageHandler.CreatePackage)
					packages.GET("", packageHandler.ListPackages)
					packages.GET("/:package", packageHandler.GetPackage)
					packages.PUT("/:package", packageHandler.UpdatePackage)
					packages.DELETE("/:package", packageHandler.DeletePackage)

					// Version routes (nested under package)
					versions := packages.Group("/:package/versions")
					{
						versions.POST("", versionHandler.PublishVersion)
						versions.GET("", versionHandler.ListVersions)
						versions.GET("/:version", versionHandler.GetVersion)
						versions.DELETE("/:version", versionHandler.DeleteVersion)
					}
				}
			}
		}
	}

	return router
}
