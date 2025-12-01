package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/criteo/command-launcher-registry/internal/apierrors"
	"github.com/criteo/command-launcher-registry/internal/models"
	"github.com/criteo/command-launcher-registry/internal/storage"
)

// VersionHandler handles version CRUD operations
type VersionHandler struct {
	store  storage.Store
	logger *slog.Logger
}

// NewVersionHandler creates a new version handler
func NewVersionHandler(store storage.Store, logger *slog.Logger) *VersionHandler {
	return &VersionHandler{
		store:  store,
		logger: logger,
	}
}

// CreateVersion handles POST /api/v1/registry/:name/package/:package/version
func (h *VersionHandler) CreateVersion(w http.ResponseWriter, r *http.Request) {
	registryName := chi.URLParam(r, "name")
	packageName := chi.URLParam(r, "package")

	var version models.Version

	// Parse request body
	if err := json.NewDecoder(r.Body).Decode(&version); err != nil {
		h.logger.Warn("Failed to decode version creation request",
			"registry", registryName,
			"package", packageName,
			"error", err,
			"remote_addr", r.RemoteAddr)
		apierrors.WriteError(w, apierrors.ErrCodeValidationError, "Invalid JSON in request body", http.StatusBadRequest, nil)
		return
	}

	// Validate version
	if err := models.ValidateVersionData(&version); err != nil {
		h.logger.Warn("Version validation failed",
			"registry", registryName,
			"package", packageName,
			"version", version.Version,
			"error", err,
			"remote_addr", r.RemoteAddr)
		apierrors.WriteError(w, apierrors.ErrCodeValidationError, err.Error(), http.StatusBadRequest, nil)
		return
	}

	// Create version
	if err := h.store.CreateVersion(r.Context(), registryName, packageName, &version); err != nil {
		if err == storage.ErrNotFound {
			// Determine what was not found
			if _, regErr := h.store.GetRegistry(r.Context(), registryName); regErr == storage.ErrNotFound {
				code, msg, status := apierrors.MapStorageError(err, "registry")
				apierrors.WriteError(w, code, msg, status, nil)
			} else {
				code, msg, status := apierrors.MapStorageError(err, "package")
				apierrors.WriteError(w, code, msg, status, nil)
			}
			return
		}
		if err == storage.ErrAlreadyExists {
			code, msg, status := apierrors.MapStorageError(err, "version")
			apierrors.WriteError(w, code, msg, status, nil)
			return
		}
		if err == storage.ErrPartitionOverlap {
			code, msg, status := apierrors.MapStorageError(err, "version")
			apierrors.WriteError(w, code, msg, status, nil)
			return
		}

		h.logger.Error("Failed to create version",
			"registry", registryName,
			"package", packageName,
			"version", version.Version,
			"error", err)
		apierrors.WriteError(w, apierrors.ErrCodeStorageUnavailable, "Failed to create version", http.StatusInternalServerError, nil)
		return
	}

	// Log successful creation
	h.logger.Info("Version created",
		"registry", registryName,
		"package", packageName,
		"version", version.Version,
		"partitions", version.StartPartition,
		"partition_end", version.EndPartition,
		"remote_addr", r.RemoteAddr)

	// Return created version
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(version)
}

// GetVersion handles GET /api/v1/registry/:name/package/:package/version/:version
func (h *VersionHandler) GetVersion(w http.ResponseWriter, r *http.Request) {
	registryName := chi.URLParam(r, "name")
	packageName := chi.URLParam(r, "package")
	versionNum := chi.URLParam(r, "version")

	// Get version from storage
	version, err := h.store.GetVersion(r.Context(), registryName, packageName, versionNum)
	if err != nil {
		if err == storage.ErrNotFound {
			// Determine what was not found
			if _, regErr := h.store.GetRegistry(r.Context(), registryName); regErr == storage.ErrNotFound {
				code, msg, status := apierrors.MapStorageError(err, "registry")
				apierrors.WriteError(w, code, msg, status, nil)
			} else if _, pkgErr := h.store.GetPackage(r.Context(), registryName, packageName); pkgErr == storage.ErrNotFound {
				code, msg, status := apierrors.MapStorageError(err, "package")
				apierrors.WriteError(w, code, msg, status, nil)
			} else {
				code, msg, status := apierrors.MapStorageError(err, "version")
				apierrors.WriteError(w, code, msg, status, nil)
			}
			return
		}

		h.logger.Error("Failed to get version",
			"registry", registryName,
			"package", packageName,
			"version", versionNum,
			"error", err)
		apierrors.WriteError(w, apierrors.ErrCodeStorageUnavailable, "Failed to retrieve version", http.StatusInternalServerError, nil)
		return
	}

	// Log retrieval
	h.logger.Debug("Version retrieved",
		"registry", registryName,
		"package", packageName,
		"version", versionNum)

	// Return version
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(version)
}

// DeleteVersion handles DELETE /api/v1/registry/:name/package/:package/version/:version
func (h *VersionHandler) DeleteVersion(w http.ResponseWriter, r *http.Request) {
	registryName := chi.URLParam(r, "name")
	packageName := chi.URLParam(r, "package")
	versionNum := chi.URLParam(r, "version")

	// Delete version
	if err := h.store.DeleteVersion(r.Context(), registryName, packageName, versionNum); err != nil {
		if err == storage.ErrNotFound {
			// Determine what was not found
			if _, regErr := h.store.GetRegistry(r.Context(), registryName); regErr == storage.ErrNotFound {
				code, msg, status := apierrors.MapStorageError(err, "registry")
				apierrors.WriteError(w, code, msg, status, nil)
			} else if _, pkgErr := h.store.GetPackage(r.Context(), registryName, packageName); pkgErr == storage.ErrNotFound {
				code, msg, status := apierrors.MapStorageError(err, "package")
				apierrors.WriteError(w, code, msg, status, nil)
			} else {
				code, msg, status := apierrors.MapStorageError(err, "version")
				apierrors.WriteError(w, code, msg, status, nil)
			}
			return
		}

		h.logger.Error("Failed to delete version",
			"registry", registryName,
			"package", packageName,
			"version", versionNum,
			"error", err)
		apierrors.WriteError(w, apierrors.ErrCodeStorageUnavailable, "Failed to delete version", http.StatusInternalServerError, nil)
		return
	}

	// Log successful deletion
	h.logger.Info("Version deleted",
		"registry", registryName,
		"package", packageName,
		"version", versionNum,
		"remote_addr", r.RemoteAddr)

	// Return 204 No Content
	w.WriteHeader(http.StatusNoContent)
}

// ListVersions handles GET /api/v1/registry/:name/package/:package/version
func (h *VersionHandler) ListVersions(w http.ResponseWriter, r *http.Request) {
	registryName := chi.URLParam(r, "name")
	packageName := chi.URLParam(r, "package")

	// Get all versions from storage
	versions, err := h.store.ListVersions(r.Context(), registryName, packageName)
	if err != nil {
		if err == storage.ErrNotFound {
			// Determine what was not found
			if _, regErr := h.store.GetRegistry(r.Context(), registryName); regErr == storage.ErrNotFound {
				code, msg, status := apierrors.MapStorageError(err, "registry")
				apierrors.WriteError(w, code, msg, status, nil)
			} else {
				code, msg, status := apierrors.MapStorageError(err, "package")
				apierrors.WriteError(w, code, msg, status, nil)
			}
			return
		}

		h.logger.Error("Failed to list versions",
			"registry", registryName,
			"package", packageName,
			"error", err)
		apierrors.WriteError(w, apierrors.ErrCodeStorageUnavailable, "Failed to list versions", http.StatusInternalServerError, nil)
		return
	}

	// Log retrieval
	h.logger.Debug("Versions listed",
		"registry", registryName,
		"package", packageName,
		"count", len(versions))

	// Return versions
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(versions)
}
