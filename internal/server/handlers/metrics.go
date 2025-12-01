package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync/atomic"
)

// MetricsHandler handles metrics requests
type MetricsHandler struct {
	logger *slog.Logger

	// Atomic counters for thread-safe increments
	totalRequests     atomic.Uint64
	indexRequests     atomic.Uint64
	registryCreates   atomic.Uint64
	registryReads     atomic.Uint64
	registryUpdates   atomic.Uint64
	registryDeletes   atomic.Uint64
	packageCreates    atomic.Uint64
	packageReads      atomic.Uint64
	packageUpdates    atomic.Uint64
	packageDeletes    atomic.Uint64
	versionCreates    atomic.Uint64
	versionReads      atomic.Uint64
	versionDeletes    atomic.Uint64
	authFailures      atomic.Uint64
	rateLimitExceeded atomic.Uint64
	validationErrors  atomic.Uint64
}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler(logger *slog.Logger) *MetricsHandler {
	return &MetricsHandler{
		logger: logger,
	}
}

// MetricsResponse represents the metrics response
type MetricsResponse struct {
	Total    uint64            `json:"total_requests"`
	ByType   map[string]uint64 `json:"by_type"`
	ByStatus map[string]uint64 `json:"by_status"`
}

// GetMetrics handles GET /api/v1/metrics
func (h *MetricsHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	response := MetricsResponse{
		Total: h.totalRequests.Load(),
		ByType: map[string]uint64{
			"index_requests":   h.indexRequests.Load(),
			"registry_creates": h.registryCreates.Load(),
			"registry_reads":   h.registryReads.Load(),
			"registry_updates": h.registryUpdates.Load(),
			"registry_deletes": h.registryDeletes.Load(),
			"package_creates":  h.packageCreates.Load(),
			"package_reads":    h.packageReads.Load(),
			"package_updates":  h.packageUpdates.Load(),
			"package_deletes":  h.packageDeletes.Load(),
			"version_creates":  h.versionCreates.Load(),
			"version_reads":    h.versionReads.Load(),
			"version_deletes":  h.versionDeletes.Load(),
		},
		ByStatus: map[string]uint64{
			"auth_failures":       h.authFailures.Load(),
			"rate_limit_exceeded": h.rateLimitExceeded.Load(),
			"validation_errors":   h.validationErrors.Load(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Request counter methods

func (h *MetricsHandler) IncrementTotalRequests() {
	h.totalRequests.Add(1)
}

func (h *MetricsHandler) IncrementIndexRequests() {
	h.indexRequests.Add(1)
}

func (h *MetricsHandler) IncrementRegistryCreates() {
	h.registryCreates.Add(1)
}

func (h *MetricsHandler) IncrementRegistryReads() {
	h.registryReads.Add(1)
}

func (h *MetricsHandler) IncrementRegistryUpdates() {
	h.registryUpdates.Add(1)
}

func (h *MetricsHandler) IncrementRegistryDeletes() {
	h.registryDeletes.Add(1)
}

func (h *MetricsHandler) IncrementPackageCreates() {
	h.packageCreates.Add(1)
}

func (h *MetricsHandler) IncrementPackageReads() {
	h.packageReads.Add(1)
}

func (h *MetricsHandler) IncrementPackageUpdates() {
	h.packageUpdates.Add(1)
}

func (h *MetricsHandler) IncrementPackageDeletes() {
	h.packageDeletes.Add(1)
}

func (h *MetricsHandler) IncrementVersionCreates() {
	h.versionCreates.Add(1)
}

func (h *MetricsHandler) IncrementVersionReads() {
	h.versionReads.Add(1)
}

func (h *MetricsHandler) IncrementVersionDeletes() {
	h.versionDeletes.Add(1)
}

func (h *MetricsHandler) IncrementAuthFailures() {
	h.authFailures.Add(1)
}

func (h *MetricsHandler) IncrementRateLimitExceeded() {
	h.rateLimitExceeded.Add(1)
}

func (h *MetricsHandler) IncrementValidationErrors() {
	h.validationErrors.Add(1)
}
