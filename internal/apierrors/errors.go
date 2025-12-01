package apierrors

import (
	"encoding/json"
	"net/http"

	"github.com/criteo/command-launcher-registry/internal/storage"
)

// ErrorCode represents standardized error codes
type ErrorCode string

const (
	ErrCodeRegistryNotFound      ErrorCode = "REGISTRY_NOT_FOUND"
	ErrCodeRegistryAlreadyExists ErrorCode = "REGISTRY_ALREADY_EXISTS"
	ErrCodePackageNotFound       ErrorCode = "PACKAGE_NOT_FOUND"
	ErrCodePackageAlreadyExists  ErrorCode = "PACKAGE_ALREADY_EXISTS"
	ErrCodeVersionNotFound       ErrorCode = "VERSION_NOT_FOUND"
	ErrCodeVersionAlreadyExists  ErrorCode = "VERSION_ALREADY_EXISTS"
	ErrCodeValidationError       ErrorCode = "VALIDATION_ERROR"
	ErrCodeInvalidPartition      ErrorCode = "INVALID_PARTITION"
	ErrCodePartitionOverlap      ErrorCode = "PARTITION_OVERLAP"
	ErrCodeStorageUnavailable    ErrorCode = "STORAGE_UNAVAILABLE"
	ErrCodeUnauthorized          ErrorCode = "UNAUTHORIZED"
)

// ErrorResponse represents the standard error response format
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error information
type ErrorDetail struct {
	Code    ErrorCode         `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

// WriteError writes a standardized error response
func WriteError(w http.ResponseWriter, code ErrorCode, message string, statusCode int, details map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
	}

	json.NewEncoder(w).Encode(response)
}

// MapStorageError maps storage errors to HTTP responses
func MapStorageError(err error, resourceType string) (ErrorCode, string, int) {
	switch err {
	case storage.ErrNotFound:
		switch resourceType {
		case "registry":
			return ErrCodeRegistryNotFound, "Registry not found", http.StatusNotFound
		case "package":
			return ErrCodePackageNotFound, "Package not found", http.StatusNotFound
		case "version":
			return ErrCodeVersionNotFound, "Version not found", http.StatusNotFound
		default:
			return ErrCodeRegistryNotFound, "Resource not found", http.StatusNotFound
		}

	case storage.ErrAlreadyExists:
		switch resourceType {
		case "registry":
			return ErrCodeRegistryAlreadyExists, "Registry already exists", http.StatusConflict
		case "package":
			return ErrCodePackageAlreadyExists, "Package already exists", http.StatusConflict
		case "version":
			return ErrCodeVersionAlreadyExists, "Version already exists (immutability violation)", http.StatusConflict
		default:
			return ErrCodeRegistryAlreadyExists, "Resource already exists", http.StatusConflict
		}

	case storage.ErrStorageUnavailable:
		return ErrCodeStorageUnavailable, "Storage service unavailable", http.StatusServiceUnavailable

	case storage.ErrImmutabilityViolation:
		return ErrCodeVersionAlreadyExists, "Version already exists (immutability violation)", http.StatusConflict

	case storage.ErrPartitionOverlap:
		return ErrCodePartitionOverlap, "Partition ranges overlap with existing version", http.StatusBadRequest

	default:
		return ErrCodeStorageUnavailable, "Internal server error", http.StatusInternalServerError
	}
}
