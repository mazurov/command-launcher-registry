package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/criteo/command-launcher-registry/internal/auth"
	"github.com/criteo/command-launcher-registry/internal/config"
	"github.com/criteo/command-launcher-registry/internal/server/middleware"
	"github.com/criteo/command-launcher-registry/internal/storage"
)

// HandlerSet contains all HTTP handlers
type HandlerSet struct {
	IndexGet     http.HandlerFunc
	IndexOptions http.HandlerFunc
	Health       http.HandlerFunc
	Metrics      http.HandlerFunc
	Whoami       http.HandlerFunc

	// Registry handlers
	ListRegistries http.HandlerFunc
	CreateRegistry http.HandlerFunc
	GetRegistry    http.HandlerFunc
	UpdateRegistry http.HandlerFunc
	DeleteRegistry http.HandlerFunc

	// Package handlers
	ListPackages  http.HandlerFunc
	CreatePackage http.HandlerFunc
	GetPackage    http.HandlerFunc
	UpdatePackage http.HandlerFunc
	DeletePackage http.HandlerFunc

	// Version handlers
	ListVersions  http.HandlerFunc
	CreateVersion http.HandlerFunc
	GetVersion    http.HandlerFunc
	DeleteVersion http.HandlerFunc
}

// Server represents the HTTP server
type Server struct {
	config        *config.Config
	logger        *slog.Logger
	store         storage.Store
	authenticator auth.Authenticator
	httpServer    *http.Server
	handlers      HandlerSet
}

// NewServer creates a new server instance
func NewServer(cfg *config.Config, logger *slog.Logger, store storage.Store, authenticator auth.Authenticator) *Server {
	return &Server{
		config:        cfg,
		logger:        logger,
		store:         store,
		authenticator: authenticator,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	// Create router
	router := s.setupRouter()

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port)
	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Log server start
	s.logger.Info("Starting server",
		"host", s.config.Server.Host,
		"port", s.config.Server.Port,
		"storage_type", s.config.Storage.Type,
		"storage_path", s.config.Storage.Path,
		"auth_type", s.config.Auth.Type)

	// Start server in goroutine
	serverErr := make(chan error, 1)
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal or server error
	select {
	case err := <-serverErr:
		return fmt.Errorf("server error: %w", err)
	case sig := <-quit:
		s.logger.Info("Shutdown signal received", "signal", sig.String())
		return s.Shutdown()
	}
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	s.logger.Info("Initiating graceful shutdown")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error("Server shutdown failed", "error", err)
		return err
	}

	// Close storage
	if err := s.store.Close(); err != nil {
		s.logger.Error("Storage close failed", "error", err)
		return err
	}

	s.logger.Info("Server stopped gracefully")
	return nil
}

// setupRouter configures the HTTP router with middleware and routes
func (s *Server) setupRouter() *chi.Mux {
	router := chi.NewRouter()

	// Global middleware (applied to all routes)
	router.Use(middleware.Logging(s.logger))
	router.Use(middleware.NewRateLimiter(100)) // 100 req/min per IP
	router.Use(middleware.CORS())

	// API v1 routes
	router.Route("/api/v1", func(r chi.Router) {
		// Health and metrics endpoints (no auth required)
		if s.handlers.Health != nil {
			r.Get("/health", s.handlers.Health)
		}
		if s.handlers.Metrics != nil {
			r.Get("/metrics", s.handlers.Metrics)
		}

		// Whoami endpoint (auth required)
		if s.handlers.Whoami != nil {
			r.Get("/whoami", s.handlers.Whoami)
		}

		// Registry index endpoint (no auth required for GET)
		r.Get("/registry/{name}/index.json", s.serveIndexPlaceholder)
		r.Options("/registry/{name}/index.json", s.handleOptionsPlaceholder)

		// Registry endpoints
		r.Route("/registry", func(r chi.Router) {
			// List registries (auth required)
			if s.handlers.ListRegistries != nil {
				r.With(middleware.RequireAuth(s.authenticator)).Get("/", s.handlers.ListRegistries)
			}

			// Create registry (auth required)
			if s.handlers.CreateRegistry != nil {
				r.With(middleware.RequireAuth(s.authenticator)).Post("/", s.handlers.CreateRegistry)
			}

			// Single registry operations
			r.Route("/{name}", func(r chi.Router) {
				// Get registry (no auth required)
				if s.handlers.GetRegistry != nil {
					r.Get("/", s.handlers.GetRegistry)
				}

				// Update registry (auth required)
				if s.handlers.UpdateRegistry != nil {
					r.With(middleware.RequireAuth(s.authenticator)).Put("/", s.handlers.UpdateRegistry)
				}

				// Delete registry (auth required)
				if s.handlers.DeleteRegistry != nil {
					r.With(middleware.RequireAuth(s.authenticator)).Delete("/", s.handlers.DeleteRegistry)
				}

				// Package endpoints
				r.Route("/package", func(r chi.Router) {
					// List packages (no auth required)
					if s.handlers.ListPackages != nil {
						r.Get("/", s.handlers.ListPackages)
					}

					// Create package (auth required)
					if s.handlers.CreatePackage != nil {
						r.With(middleware.RequireAuth(s.authenticator)).Post("/", s.handlers.CreatePackage)
					}

					// Single package operations
					r.Route("/{package}", func(r chi.Router) {
						// Get package (no auth required)
						if s.handlers.GetPackage != nil {
							r.Get("/", s.handlers.GetPackage)
						}

						// Update package (auth required)
						if s.handlers.UpdatePackage != nil {
							r.With(middleware.RequireAuth(s.authenticator)).Put("/", s.handlers.UpdatePackage)
						}

						// Delete package (auth required)
						if s.handlers.DeletePackage != nil {
							r.With(middleware.RequireAuth(s.authenticator)).Delete("/", s.handlers.DeletePackage)
						}

						// Version endpoints
						r.Route("/version", func(r chi.Router) {
							// List versions (no auth required)
							if s.handlers.ListVersions != nil {
								r.Get("/", s.handlers.ListVersions)
							}

							// Create version (auth required)
							if s.handlers.CreateVersion != nil {
								r.With(middleware.RequireAuth(s.authenticator)).Post("/", s.handlers.CreateVersion)
							}

							// Single version operations
							r.Route("/{version}", func(r chi.Router) {
								// Get version (no auth required)
								if s.handlers.GetVersion != nil {
									r.Get("/", s.handlers.GetVersion)
								}

								// Delete version (auth required)
								if s.handlers.DeleteVersion != nil {
									r.With(middleware.RequireAuth(s.authenticator)).Delete("/", s.handlers.DeleteVersion)
								}
							})
						})
					})
				})
			})
		})
	})

	return router
}

// SetHandlers sets all handlers (called from main to avoid import cycle)
func (s *Server) SetHandlers(handlers HandlerSet) {
	s.handlers = handlers
}

func (s *Server) serveIndexPlaceholder(w http.ResponseWriter, r *http.Request) {
	if s.handlers.IndexGet != nil {
		s.handlers.IndexGet(w, r)
	} else {
		http.Error(w, "Index handler not configured", http.StatusInternalServerError)
	}
}

func (s *Server) handleOptionsPlaceholder(w http.ResponseWriter, r *http.Request) {
	if s.handlers.IndexOptions != nil {
		s.handlers.IndexOptions(w, r)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}
