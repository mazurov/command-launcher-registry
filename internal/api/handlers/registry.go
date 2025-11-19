package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mazurov/command-launcher-registry/internal/repository"
	"github.com/mazurov/command-launcher-registry/internal/service"
	"github.com/mazurov/command-launcher-registry/pkg/types"
)

type RegistryHandler struct {
	service *service.Service
}

func NewRegistryHandler(svc *service.Service) *RegistryHandler {
	return &RegistryHandler{service: svc}
}

// CreateRegistry godoc
// @Summary Create a new registry
// @Tags registries
// @Accept json
// @Produce json
// @Param registry body types.CreateRegistryRequest true "Registry data"
// @Success 201 {object} models.Registry
// @Failure 400 {object} types.ErrorResponse
// @Failure 409 {object} types.ErrorResponse
// @Router /remote/registries [post]
func (h *RegistryHandler) CreateRegistry(c *gin.Context) {
	var req types.CreateRegistryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	registry, err := h.service.CreateRegistry(&req)
	if err != nil {
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

	c.JSON(http.StatusCreated, registry)
}

// GetRegistry godoc
// @Summary Get a registry by name
// @Tags registries
// @Produce json
// @Param name path string true "Registry name"
// @Success 200 {object} models.Registry
// @Failure 404 {object} types.ErrorResponse
// @Router /remote/registries/{name} [get]
func (h *RegistryHandler) GetRegistry(c *gin.Context) {
	name := c.Param("name")

	registry, err := h.service.GetRegistry(name)
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

	c.JSON(http.StatusOK, registry)
}

// ListRegistries godoc
// @Summary List all registries
// @Tags registries
// @Produce json
// @Success 200 {array} models.Registry
// @Router /remote/registries [get]
func (h *RegistryHandler) ListRegistries(c *gin.Context) {
	registries, err := h.service.ListRegistries()
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, registries)
}

// UpdateRegistry godoc
// @Summary Update a registry
// @Tags registries
// @Accept json
// @Produce json
// @Param name path string true "Registry name"
// @Param registry body types.UpdateRegistryRequest true "Update data"
// @Success 200 {object} types.SuccessResponse
// @Failure 404 {object} types.ErrorResponse
// @Router /remote/registries/{name} [put]
func (h *RegistryHandler) UpdateRegistry(c *gin.Context) {
	name := c.Param("name")
	var req types.UpdateRegistryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	if err := h.service.UpdateRegistry(name, &req); err != nil {
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

	c.JSON(http.StatusOK, types.SuccessResponse{
		Success: true,
		Message: "Registry updated successfully",
	})
}

// DeleteRegistry godoc
// @Summary Delete a registry
// @Tags registries
// @Param name path string true "Registry name"
// @Success 200 {object} types.SuccessResponse
// @Failure 404 {object} types.ErrorResponse
// @Router /remote/registries/{name} [delete]
func (h *RegistryHandler) DeleteRegistry(c *gin.Context) {
	name := c.Param("name")

	if err := h.service.DeleteRegistry(name); err != nil {
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

	c.JSON(http.StatusOK, types.SuccessResponse{
		Success: true,
		Message: "Registry deleted successfully",
	})
}

// GetRegistryIndex godoc
// @Summary Get CDT-compatible registry index
// @Description Returns a flat list of all package versions in the registry
// @Tags compatibility
// @Produce json
// @Param name path string true "Registry name"
// @Success 200 {array} types.PackageInfo
// @Failure 404 {object} types.ErrorResponse
// @Router /registry/{name}/index.json [get]
func (h *RegistryHandler) GetRegistryIndex(c *gin.Context) {
	name := c.Param("name")

	index, err := h.service.GetRegistryIndex(name)
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

	c.JSON(http.StatusOK, index)
}
