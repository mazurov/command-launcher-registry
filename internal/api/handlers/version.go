package handlers

import (
	"errors"
	"net/http"

	"github.com/mazurov/command-launcher-registry/internal/repository"
	"github.com/mazurov/command-launcher-registry/internal/service"
	"github.com/mazurov/command-launcher-registry/pkg/types"
	"github.com/gin-gonic/gin"
)

type VersionHandler struct {
	service *service.Service
}

func NewVersionHandler(svc *service.Service) *VersionHandler {
	return &VersionHandler{service: svc}
}

// PublishVersion godoc
// @Summary Publish a new version
// @Tags versions
// @Accept json
// @Produce json
// @Param registry path string true "Registry name"
// @Param package path string true "Package name"
// @Param version body types.PublishVersionRequest true "Version data"
// @Success 201 {object} models.Version
// @Failure 400 {object} types.ErrorResponse
// @Failure 404 {object} types.ErrorResponse
// @Failure 409 {object} types.ErrorResponse
// @Router /remote/registries/{registry}/packages/{package}/versions [post]
func (h *VersionHandler) PublishVersion(c *gin.Context) {
	registryName := c.Param("name")
	packageName := c.Param("package")
	var req types.PublishVersionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	version, err := h.service.PublishVersion(registryName, packageName, &req)
	if err != nil {
		if errors.Is(err, repository.ErrRegistryNotFound) || errors.Is(err, repository.ErrPackageNotFound) {
			c.JSON(http.StatusNotFound, types.ErrorResponse{
				Error:   "not_found",
				Message: err.Error(),
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

	c.JSON(http.StatusCreated, version)
}

// GetVersion godoc
// @Summary Get a specific version
// @Tags versions
// @Produce json
// @Param registry path string true "Registry name"
// @Param package path string true "Package name"
// @Param version path string true "Version string"
// @Success 200 {object} models.Version
// @Failure 404 {object} types.ErrorResponse
// @Router /remote/registries/{registry}/packages/{package}/versions/{version} [get]
func (h *VersionHandler) GetVersion(c *gin.Context) {
	registryName := c.Param("name")
	packageName := c.Param("package")
	versionStr := c.Param("version")

	version, err := h.service.GetVersion(registryName, packageName, versionStr)
	if err != nil {
		if errors.Is(err, repository.ErrVersionNotFound) || errors.Is(err, repository.ErrPackageNotFound) {
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

	c.JSON(http.StatusOK, version)
}

// ListVersions godoc
// @Summary List all versions of a package
// @Tags versions
// @Produce json
// @Param registry path string true "Registry name"
// @Param package path string true "Package name"
// @Success 200 {array} models.Version
// @Failure 404 {object} types.ErrorResponse
// @Router /remote/registries/{registry}/packages/{package}/versions [get]
func (h *VersionHandler) ListVersions(c *gin.Context) {
	registryName := c.Param("name")
	packageName := c.Param("package")

	versions, err := h.service.ListVersions(registryName, packageName)
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

	c.JSON(http.StatusOK, versions)
}

// DeleteVersion godoc
// @Summary Delete a version
// @Tags versions
// @Param registry path string true "Registry name"
// @Param package path string true "Package name"
// @Param version path string true "Version string"
// @Success 200 {object} types.SuccessResponse
// @Failure 404 {object} types.ErrorResponse
// @Router /remote/registries/{registry}/packages/{package}/versions/{version} [delete]
func (h *VersionHandler) DeleteVersion(c *gin.Context) {
	registryName := c.Param("name")
	packageName := c.Param("package")
	versionStr := c.Param("version")

	if err := h.service.DeleteVersion(registryName, packageName, versionStr); err != nil {
		if errors.Is(err, repository.ErrVersionNotFound) || errors.Is(err, repository.ErrPackageNotFound) {
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
		Message: "Version deleted successfully",
	})
}
