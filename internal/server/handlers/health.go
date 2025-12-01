package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/criteo/command-launcher-registry/internal/storage"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	store  storage.Store
	logger *slog.Logger
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(store storage.Store, logger *slog.Logger) *HealthHandler {
	return &HealthHandler{
		store:  store,
		logger: logger,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status string                 `json:"status"`
	Checks map[string]CheckResult `json:"checks"`
}

// CheckResult represents a single health check result
type CheckResult struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// GetHealth handles GET /api/v1/health
func (h *HealthHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status: "healthy",
		Checks: make(map[string]CheckResult),
	}

	// Check storage connectivity
	// We use ListRegistries as a basic connectivity check
	_, err := h.store.ListRegistries(r.Context())
	if err != nil {
		response.Checks["storage"] = CheckResult{
			Status:  "unhealthy",
			Message: err.Error(),
		}
		response.Status = "unhealthy"

		h.logger.Error("Health check failed: storage unhealthy", "error", err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(response)
		return
	}

	response.Checks["storage"] = CheckResult{
		Status: "healthy",
	}

	// Return healthy response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
