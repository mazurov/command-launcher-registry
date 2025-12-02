package storage

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
)

// OCI error categories for clear error messages
const (
	OCICategoryAuth    = "authentication"
	OCICategoryNetwork = "network"
	OCICategoryStorage = "storage"
)

// OCI operations for error context
const (
	OCIOpPush    = "push"
	OCIOpPull    = "pull"
	OCIOpConnect = "connect"
)

// OCIError wraps OCI-specific failures with categorization
type OCIError struct {
	Category string // "authentication", "network", or "storage"
	Op       string // "push", "pull", or "connect"
	Err      error  // Underlying error
}

// Error implements the error interface
func (e *OCIError) Error() string {
	return fmt.Sprintf("OCI %s error during %s: %v", e.Category, e.Op, e.Err)
}

// Unwrap implements the errors.Unwrap interface
func (e *OCIError) Unwrap() error {
	return e.Err
}

// Is implements the errors.Is interface to match ErrStorageUnavailable
func (e *OCIError) Is(target error) bool {
	return target == ErrStorageUnavailable
}

// NewOCIAuthError creates an authentication-related OCI error
func NewOCIAuthError(op string, err error) *OCIError {
	return &OCIError{
		Category: OCICategoryAuth,
		Op:       op,
		Err:      err,
	}
}

// NewOCINetworkError creates a network-related OCI error
func NewOCINetworkError(op string, err error) *OCIError {
	return &OCIError{
		Category: OCICategoryNetwork,
		Op:       op,
		Err:      err,
	}
}

// NewOCIStorageError creates a storage-related OCI error
func NewOCIStorageError(op string, err error) *OCIError {
	return &OCIError{
		Category: OCICategoryStorage,
		Op:       op,
		Err:      err,
	}
}

// CategorizeOCIError examines an error and returns an appropriately categorized OCIError.
// It checks for HTTP status codes, network errors, and other common failure patterns.
// Includes registry-specific hints for common authentication issues.
func CategorizeOCIError(op string, err error) *OCIError {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	// Check for authentication errors (401, 403)
	if containsHTTPStatus(errStr, 401) || strings.Contains(errStr, "UNAUTHORIZED") {
		hint := getRegistryAuthHint(errStr)
		return NewOCIAuthError(op, fmt.Errorf("authentication failed: verify storage token is valid%s", hint))
	}
	if containsHTTPStatus(errStr, 403) || strings.Contains(errStr, "FORBIDDEN") {
		hint := getRegistryAuthHint(errStr)
		return NewOCIAuthError(op, fmt.Errorf("access denied: token lacks required permissions%s", hint))
	}

	// Check for network errors
	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return NewOCINetworkError(op, fmt.Errorf("network error: unable to reach OCI registry (timeout)"))
		}
		return NewOCINetworkError(op, fmt.Errorf("network error: unable to reach OCI registry"))
	}

	// Check for DNS errors
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return NewOCINetworkError(op, fmt.Errorf("network error: cannot resolve registry hostname"))
	}

	// Check for URL errors (connection refused, etc.)
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		if urlErr.Timeout() {
			return NewOCINetworkError(op, fmt.Errorf("network error: unable to reach OCI registry (timeout)"))
		}
		// Connection refused, no route to host, etc.
		return NewOCINetworkError(op, fmt.Errorf("network error: unable to reach OCI registry"))
	}

	// Check for storage errors (404, 500, 503)
	if containsHTTPStatus(errStr, 404) || strings.Contains(errStr, "NOT_FOUND") {
		return NewOCIStorageError(op, fmt.Errorf("repository not found or not initialized"))
	}
	if containsHTTPStatus(errStr, 500) || containsHTTPStatus(errStr, 503) {
		return NewOCIStorageError(op, fmt.Errorf("OCI registry unavailable: %v", err))
	}

	// Default to storage error
	return NewOCIStorageError(op, err)
}

// containsHTTPStatus checks if the error string contains a specific HTTP status code
func containsHTTPStatus(errStr string, status int) bool {
	// Check for common patterns like "401", "status 401", "status: 401", "HTTP 401"
	patterns := []string{
		fmt.Sprintf("%d", status),
		fmt.Sprintf("status %d", status),
		fmt.Sprintf("status: %d", status),
		fmt.Sprintf("HTTP %d", status),
	}
	for _, pattern := range patterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}
	return false
}

// getRegistryAuthHint returns registry-specific authentication hints based on the error
func getRegistryAuthHint(errStr string) string {
	errLower := strings.ToLower(errStr)

	// GitHub Container Registry (ghcr.io)
	if strings.Contains(errLower, "ghcr.io") {
		return " (ghcr.io: use a GitHub PAT with 'write:packages' scope)"
	}

	// Docker Hub (docker.io, index.docker.io)
	if strings.Contains(errLower, "docker.io") || strings.Contains(errLower, "index.docker.io") {
		return " (docker.io: use a Docker Hub access token, may require username:token format)"
	}

	// Azure Container Registry (*.azurecr.io)
	if strings.Contains(errLower, "azurecr.io") {
		return " (Azure ACR: use 'az acr login --expose-token' to get a token)"
	}

	// Amazon ECR (*.amazonaws.com)
	if strings.Contains(errLower, "amazonaws.com") || strings.Contains(errLower, "ecr") {
		return " (AWS ECR: use 'aws ecr get-login-password' to get a token)"
	}

	// Google Container Registry / Artifact Registry
	if strings.Contains(errLower, "gcr.io") || strings.Contains(errLower, "pkg.dev") {
		return " (GCP: use 'gcloud auth print-access-token' to get a token)"
	}

	return ""
}
