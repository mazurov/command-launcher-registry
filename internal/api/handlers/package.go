package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mazurov/command-launcher-registry/internal/repository"
	"github.com/mazurov/command-launcher-registry/internal/service"
	"github.com/mazurov/command-launcher-registry/pkg/types"
)

type PackageHandler struct {
	service *service.Service
}

func NewPackageHandler(svc *service.Service) *PackageHandler {
	return &PackageHandler{service: svc}
}

// CreatePackage godoc
// @Summary Create a new package
// @Tags packages
// @Accept json
// @Produce json
// @Param registry path string true "Registry name"
// @Param package body types.CreatePackageRequest true "Package data"
// @Success 201 {object} models.Package
// @Failure 400 {object} types.ErrorResponse
// @Failure 404 {object} types.ErrorResponse
// @Failure 409 {object} types.ErrorResponse
// @Router /remote/registries/{registry}/packages [post]
func (h *PackageHandler) CreatePackage(c *gin.Context) {
	registryName := c.Param("name")
	var req types.CreatePackageRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	pkg, err := h.service.CreatePackage(registryName, &req)
	if err != nil {
		if errors.Is(err, repository.ErrRegistryNotFound) {
			c.JSON(http.StatusNotFound, types.ErrorResponse{
				Error:   "not_found",
				Message: "Registry not found",
			})
			return
		}
		if errors.Is(err, repository.ErrAlreadyExists) {
			c.JSON(http.StatusConflict, types.ErrorResponse{
				Error:   "already_exists",
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, pkg)
}

// GetPackage godoc
// @Summary Get a package
// @Tags packages
// @Produce json
// @Param registry path string true "Registry name"
// @Param package path string true "Package name"
// @Success 200 {object} models.Package
// @Failure 404 {object} types.ErrorResponse
// @Router /remote/registries/{registry}/packages/{package} [get]
func (h *PackageHandler) GetPackage(c *gin.Context) {
	registryName := c.Param("name")
	packageName := c.Param("package")

	pkg, err := h.service.GetPackage(registryName, packageName)
	if err != nil {
		if errors.Is(err, repository.ErrRegistryNotFound) || errors.Is(err, repository.ErrPackageNotFound) {
			c.JSON(http.StatusNotFound, types.ErrorResponse{
				Error:   "not_found",
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, pkg)
}

// ListPackages godoc
// @Summary List all packages in a registry
// @Tags packages
// @Produce json
// @Param registry path string true "Registry name"
// @Success 200 {array} models.Package
// @Failure 404 {object} types.ErrorResponse
// @Router /remote/registries/{registry}/packages [get]
func (h *PackageHandler) ListPackages(c *gin.Context) {
	registryName := c.Param("name")

	packages, err := h.service.ListPackages(registryName)
	if err != nil {
		if errors.Is(err, repository.ErrRegistryNotFound) {
			c.JSON(http.StatusNotFound, types.ErrorResponse{
				Error:   "not_found",
				Message: "Registry not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, packages)
}

// UpdatePackage godoc
// @Summary Update a package
// @Tags packages
// @Accept json
// @Produce json
// @Param registry path string true "Registry name"
// @Param package path string true "Package name"
// @Param data body types.UpdatePackageRequest true "Update data"
// @Success 200 {object} types.SuccessResponse
// @Failure 404 {object} types.ErrorResponse
// @Router /remote/registries/{registry}/packages/{package} [put]
func (h *PackageHandler) UpdatePackage(c *gin.Context) {
	registryName := c.Param("name")
	packageName := c.Param("package")
	var req types.UpdatePackageRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	if err := h.service.UpdatePackage(registryName, packageName, &req); err != nil {
		if errors.Is(err, repository.ErrRegistryNotFound) || errors.Is(err, repository.ErrPackageNotFound) {
			c.JSON(http.StatusNotFound, types.ErrorResponse{
				Error:   "not_found",
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, types.SuccessResponse{
		Success: true,
		Message: "Package updated successfully",
	})
}

// DeletePackage godoc
// @Summary Delete a package
// @Tags packages
// @Param registry path string true "Registry name"
// @Param package path string true "Package name"
// @Success 200 {object} types.SuccessResponse
// @Failure 404 {object} types.ErrorResponse
// @Router /remote/registries/{registry}/packages/{package} [delete]
func (h *PackageHandler) DeletePackage(c *gin.Context) {
	registryName := c.Param("name")
	packageName := c.Param("package")

	if err := h.service.DeletePackage(registryName, packageName); err != nil {
		if errors.Is(err, repository.ErrRegistryNotFound) || errors.Is(err, repository.ErrPackageNotFound) {
			c.JSON(http.StatusNotFound, types.ErrorResponse{
				Error:   "not_found",
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, types.SuccessResponse{
		Success: true,
		Message: "Package deleted successfully",
	})
}
