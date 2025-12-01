package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/criteo/command-launcher-registry/internal/apierrors"
	"github.com/criteo/command-launcher-registry/internal/storage"
)

// IndexHandler handles registry index.json requests
type IndexHandler struct {
	store  storage.Store
	logger *slog.Logger
}

// NewIndexHandler creates a new index handler
func NewIndexHandler(store storage.Store, logger *slog.Logger) *IndexHandler {
	return &IndexHandler{
		store:  store,
		logger: logger,
	}
}

// GetIndex handles GET /api/v1/registry/:name/index.json
func (h *IndexHandler) GetIndex(w http.ResponseWriter, r *http.Request) {
	registryName := chi.URLParam(r, "name")

	// Get registry index from storage
	entries, err := h.store.GetRegistryIndex(r.Context(), registryName)
	if err != nil {
		if err == storage.ErrNotFound {
			code, msg, status := apierrors.MapStorageError(err, "registry")
			apierrors.WriteError(w, code, msg, status, nil)
			return
		}

		h.logger.Error("Failed to get registry index",
			"registry", registryName,
			"error", err)
		apierrors.WriteError(w, apierrors.ErrCodeStorageUnavailable, "Failed to retrieve index", http.StatusInternalServerError, nil)
		return
	}

	// Log index request
	h.logger.Info("Registry index served",
		"registry", registryName,
		"entry_count", len(entries))

	// Return JSON array
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(entries)
}

// HandleOptions handles OPTIONS /api/v1/registry/:name/index.json (CORS preflight)
func (h *IndexHandler) HandleOptions(w http.ResponseWriter, r *http.Request) {
	// CORS headers are set by middleware
	// Just return 200 OK
	w.WriteHeader(http.StatusOK)
}
