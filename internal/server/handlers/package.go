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

// PackageHandler handles package CRUD operations
type PackageHandler struct {
	store  storage.Store
	logger *slog.Logger
}

// NewPackageHandler creates a new package handler
func NewPackageHandler(store storage.Store, logger *slog.Logger) *PackageHandler {
	return &PackageHandler{
		store:  store,
		logger: logger,
	}
}

// CreatePackage handles POST /api/v1/registry/:name/package
func (h *PackageHandler) CreatePackage(w http.ResponseWriter, r *http.Request) {
	registryName := chi.URLParam(r, "name")

	var pkg models.Package

	// Parse request body
	if err := json.NewDecoder(r.Body).Decode(&pkg); err != nil {
		h.logger.Warn("Failed to decode package creation request",
			"registry", registryName,
			"error", err,
			"remote_addr", r.RemoteAddr)
		apierrors.WriteError(w, apierrors.ErrCodeValidationError, "Invalid JSON in request body", http.StatusBadRequest, nil)
		return
	}

	// Validate package
	if err := models.ValidatePackage(&pkg); err != nil {
		h.logger.Warn("Package validation failed",
			"registry", registryName,
			"package", pkg.Name,
			"error", err,
			"remote_addr", r.RemoteAddr)
		apierrors.WriteError(w, apierrors.ErrCodeValidationError, err.Error(), http.StatusBadRequest, nil)
		return
	}

	// Initialize versions map if nil
	if pkg.Versions == nil {
		pkg.Versions = make(map[string]*models.Version)
	}

	// Create package
	if err := h.store.CreatePackage(r.Context(), registryName, &pkg); err != nil {
		if err == storage.ErrNotFound {
			code, msg, status := apierrors.MapStorageError(err, "registry")
			apierrors.WriteError(w, code, msg, status, nil)
			return
		}
		if err == storage.ErrAlreadyExists {
			code, msg, status := apierrors.MapStorageError(err, "package")
			apierrors.WriteError(w, code, msg, status, nil)
			return
		}

		h.logger.Error("Failed to create package",
			"registry", registryName,
			"package", pkg.Name,
			"error", err)
		apierrors.WriteError(w, apierrors.ErrCodeStorageUnavailable, "Failed to create package", http.StatusInternalServerError, nil)
		return
	}

	// Log successful creation
	h.logger.Info("Package created",
		"registry", registryName,
		"package", pkg.Name,
		"maintainer_count", len(pkg.Maintainers),
		"custom_values", len(pkg.CustomValues),
		"remote_addr", r.RemoteAddr)

	// Return created package
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(pkg)
}

// GetPackage handles GET /api/v1/registry/:name/package/:package
func (h *PackageHandler) GetPackage(w http.ResponseWriter, r *http.Request) {
	registryName := chi.URLParam(r, "name")
	packageName := chi.URLParam(r, "package")

	// Get package from storage
	pkg, err := h.store.GetPackage(r.Context(), registryName, packageName)
	if err != nil {
		if err == storage.ErrNotFound {
			// Determine if registry or package not found
			if _, regErr := h.store.GetRegistry(r.Context(), registryName); regErr == storage.ErrNotFound {
				code, msg, status := apierrors.MapStorageError(err, "registry")
				apierrors.WriteError(w, code, msg, status, nil)
			} else {
				code, msg, status := apierrors.MapStorageError(err, "package")
				apierrors.WriteError(w, code, msg, status, nil)
			}
			return
		}

		h.logger.Error("Failed to get package",
			"registry", registryName,
			"package", packageName,
			"error", err)
		apierrors.WriteError(w, apierrors.ErrCodeStorageUnavailable, "Failed to retrieve package", http.StatusInternalServerError, nil)
		return
	}

	// Log retrieval
	h.logger.Debug("Package retrieved",
		"registry", registryName,
		"package", packageName,
		"version_count", len(pkg.Versions))

	// Return package
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(pkg)
}

// UpdatePackage handles PUT /api/v1/registry/:name/package/:package
func (h *PackageHandler) UpdatePackage(w http.ResponseWriter, r *http.Request) {
	registryName := chi.URLParam(r, "name")
	packageName := chi.URLParam(r, "package")

	var pkg models.Package

	// Parse request body
	if err := json.NewDecoder(r.Body).Decode(&pkg); err != nil {
		h.logger.Warn("Failed to decode package update request",
			"registry", registryName,
			"package", packageName,
			"error", err,
			"remote_addr", r.RemoteAddr)
		apierrors.WriteError(w, apierrors.ErrCodeValidationError, "Invalid JSON in request body", http.StatusBadRequest, nil)
		return
	}

	// Ensure name in URL matches name in body
	if pkg.Name != packageName {
		h.logger.Warn("Package name mismatch",
			"url_name", packageName,
			"body_name", pkg.Name,
			"remote_addr", r.RemoteAddr)
		apierrors.WriteError(w, apierrors.ErrCodeValidationError, "Package name in URL must match name in body", http.StatusBadRequest, nil)
		return
	}

	// Validate package
	if err := models.ValidatePackage(&pkg); err != nil {
		h.logger.Warn("Package validation failed",
			"registry", registryName,
			"package", pkg.Name,
			"error", err,
			"remote_addr", r.RemoteAddr)
		apierrors.WriteError(w, apierrors.ErrCodeValidationError, err.Error(), http.StatusBadRequest, nil)
		return
	}

	// Get existing package to preserve versions
	existing, err := h.store.GetPackage(r.Context(), registryName, packageName)
	if err != nil {
		if err == storage.ErrNotFound {
			// Determine if registry or package not found
			if _, regErr := h.store.GetRegistry(r.Context(), registryName); regErr == storage.ErrNotFound {
				code, msg, status := apierrors.MapStorageError(err, "registry")
				apierrors.WriteError(w, code, msg, status, nil)
			} else {
				code, msg, status := apierrors.MapStorageError(err, "package")
				apierrors.WriteError(w, code, msg, status, nil)
			}
			return
		}

		h.logger.Error("Failed to get existing package",
			"registry", registryName,
			"package", packageName,
			"error", err)
		apierrors.WriteError(w, apierrors.ErrCodeStorageUnavailable, "Failed to retrieve package", http.StatusInternalServerError, nil)
		return
	}

	// Preserve versions from existing package
	pkg.Versions = existing.Versions

	// Update package
	if err := h.store.UpdatePackage(r.Context(), registryName, &pkg); err != nil {
		if err == storage.ErrNotFound {
			// Determine if registry or package not found
			if _, regErr := h.store.GetRegistry(r.Context(), registryName); regErr == storage.ErrNotFound {
				code, msg, status := apierrors.MapStorageError(err, "registry")
				apierrors.WriteError(w, code, msg, status, nil)
			} else {
				code, msg, status := apierrors.MapStorageError(err, "package")
				apierrors.WriteError(w, code, msg, status, nil)
			}
			return
		}

		h.logger.Error("Failed to update package",
			"registry", registryName,
			"package", packageName,
			"error", err)
		apierrors.WriteError(w, apierrors.ErrCodeStorageUnavailable, "Failed to update package", http.StatusInternalServerError, nil)
		return
	}

	// Log successful update
	h.logger.Info("Package updated",
		"registry", registryName,
		"package", pkg.Name,
		"maintainer_count", len(pkg.Maintainers),
		"custom_values", len(pkg.CustomValues),
		"remote_addr", r.RemoteAddr)

	// Return updated package
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(pkg)
}

// DeletePackage handles DELETE /api/v1/registry/:name/package/:package
func (h *PackageHandler) DeletePackage(w http.ResponseWriter, r *http.Request) {
	registryName := chi.URLParam(r, "name")
	packageName := chi.URLParam(r, "package")

	// Delete package (cascade delete handled by storage layer)
	if err := h.store.DeletePackage(r.Context(), registryName, packageName); err != nil {
		if err == storage.ErrNotFound {
			// Determine if registry or package not found
			if _, regErr := h.store.GetRegistry(r.Context(), registryName); regErr == storage.ErrNotFound {
				code, msg, status := apierrors.MapStorageError(err, "registry")
				apierrors.WriteError(w, code, msg, status, nil)
			} else {
				code, msg, status := apierrors.MapStorageError(err, "package")
				apierrors.WriteError(w, code, msg, status, nil)
			}
			return
		}

		h.logger.Error("Failed to delete package",
			"registry", registryName,
			"package", packageName,
			"error", err)
		apierrors.WriteError(w, apierrors.ErrCodeStorageUnavailable, "Failed to delete package", http.StatusInternalServerError, nil)
		return
	}

	// Log successful deletion
	h.logger.Info("Package deleted",
		"registry", registryName,
		"package", packageName,
		"remote_addr", r.RemoteAddr)

	// Return 204 No Content
	w.WriteHeader(http.StatusNoContent)
}

// ListPackages handles GET /api/v1/registry/:name/package
func (h *PackageHandler) ListPackages(w http.ResponseWriter, r *http.Request) {
	registryName := chi.URLParam(r, "name")

	// Get all packages from storage
	packages, err := h.store.ListPackages(r.Context(), registryName)
	if err != nil {
		if err == storage.ErrNotFound {
			code, msg, status := apierrors.MapStorageError(err, "registry")
			apierrors.WriteError(w, code, msg, status, nil)
			return
		}

		h.logger.Error("Failed to list packages",
			"registry", registryName,
			"error", err)
		apierrors.WriteError(w, apierrors.ErrCodeStorageUnavailable, "Failed to list packages", http.StatusInternalServerError, nil)
		return
	}

	// Log retrieval
	h.logger.Debug("Packages listed",
		"registry", registryName,
		"count", len(packages))

	// Return packages
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(packages)
}
