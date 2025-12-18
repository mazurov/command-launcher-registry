package storage

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/minio/minio-go/v7"
)

// S3 error categories for clear error messages
const (
	S3CategoryAuth    = "authentication"
	S3CategoryNetwork = "network"
	S3CategoryStorage = "storage"
)

// S3 operations for error context
const (
	S3OpUpload   = "upload"
	S3OpDownload = "download"
	S3OpConnect  = "connect"
)

// S3Error wraps S3-specific failures with categorization
type S3Error struct {
	Category string // "authentication", "network", or "storage"
	Op       string // "upload", "download", or "connect"
	Err      error  // Underlying error
}

// Error implements the error interface
func (e *S3Error) Error() string {
	return fmt.Sprintf("S3 %s error during %s: %v", e.Category, e.Op, e.Err)
}

// Unwrap implements the errors.Unwrap interface
func (e *S3Error) Unwrap() error {
	return e.Err
}

// Is implements the errors.Is interface to match ErrStorageUnavailable
func (e *S3Error) Is(target error) bool {
	return target == ErrStorageUnavailable
}

// NewS3AuthError creates an authentication-related S3 error
func NewS3AuthError(op string, err error) *S3Error {
	return &S3Error{
		Category: S3CategoryAuth,
		Op:       op,
		Err:      err,
	}
}

// NewS3NetworkError creates a network-related S3 error
func NewS3NetworkError(op string, err error) *S3Error {
	return &S3Error{
		Category: S3CategoryNetwork,
		Op:       op,
		Err:      err,
	}
}

// NewS3StorageError creates a storage-related S3 error
func NewS3StorageError(op string, err error) *S3Error {
	return &S3Error{
		Category: S3CategoryStorage,
		Op:       op,
		Err:      err,
	}
}

// CategorizeS3Error examines an error and returns an appropriately categorized S3Error.
// It checks for MinIO error responses, network errors, and other common failure patterns.
// Includes provider-specific hints for common authentication issues.
func CategorizeS3Error(op string, err error) *S3Error {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	// Check for MinIO ErrorResponse
	var minioErr minio.ErrorResponse
	if errors.As(err, &minioErr) {
		return categorizeMinioError(op, minioErr)
	}

	// Check for authentication errors by string patterns
	if strings.Contains(errStr, "AccessDenied") ||
		strings.Contains(errStr, "InvalidAccessKeyId") ||
		strings.Contains(errStr, "SignatureDoesNotMatch") ||
		strings.Contains(errStr, "ExpiredToken") {
		hint := getS3ProviderAuthHint(errStr)
		return NewS3AuthError(op, fmt.Errorf("authentication failed: %v%s", err, hint))
	}

	// Check for network errors
	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return NewS3NetworkError(op, fmt.Errorf("network timeout: unable to reach S3 endpoint"))
		}
		return NewS3NetworkError(op, fmt.Errorf("network error: unable to reach S3 endpoint"))
	}

	// Check for DNS errors
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return NewS3NetworkError(op, fmt.Errorf("network error: cannot resolve S3 endpoint hostname"))
	}

	// Check for URL errors (connection refused, etc.)
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		if urlErr.Timeout() {
			return NewS3NetworkError(op, fmt.Errorf("network timeout: unable to reach S3 endpoint"))
		}
		return NewS3NetworkError(op, fmt.Errorf("network error: unable to reach S3 endpoint"))
	}

	// Check for storage errors by string patterns
	if strings.Contains(errStr, "NoSuchBucket") {
		return NewS3StorageError(op, fmt.Errorf("bucket not found: verify bucket exists and name is correct"))
	}
	if strings.Contains(errStr, "NoSuchKey") {
		return NewS3StorageError(op, fmt.Errorf("object not found"))
	}

	// Default to storage error
	return NewS3StorageError(op, err)
}

// categorizeMinioError handles MinIO-specific error responses
func categorizeMinioError(op string, minioErr minio.ErrorResponse) *S3Error {
	switch minioErr.Code {
	case "AccessDenied":
		hint := getS3ProviderAuthHint(minioErr.BucketName)
		return NewS3AuthError(op, fmt.Errorf("access denied: token lacks required permissions%s", hint))
	case "InvalidAccessKeyId":
		return NewS3AuthError(op, fmt.Errorf("invalid access key: verify credentials are correct"))
	case "SignatureDoesNotMatch":
		return NewS3AuthError(op, fmt.Errorf("signature mismatch: verify secret key is correct"))
	case "ExpiredToken":
		return NewS3AuthError(op, fmt.Errorf("token expired: refresh credentials"))
	case "NoSuchBucket":
		return NewS3StorageError(op, fmt.Errorf("bucket not found: verify bucket exists and name is correct"))
	case "NoSuchKey":
		return NewS3StorageError(op, fmt.Errorf("object not found"))
	case "InternalError", "ServiceUnavailable":
		return NewS3StorageError(op, fmt.Errorf("S3 service unavailable: %s", minioErr.Message))
	default:
		return NewS3StorageError(op, fmt.Errorf("%s: %s", minioErr.Code, minioErr.Message))
	}
}

// getS3ProviderAuthHint returns provider-specific authentication hints based on the endpoint
func getS3ProviderAuthHint(endpoint string) string {
	endpointLower := strings.ToLower(endpoint)

	// AWS S3
	if strings.Contains(endpointLower, "amazonaws.com") || strings.Contains(endpointLower, "s3.") {
		return " (AWS S3: check IAM policy has s3:GetObject, s3:PutObject, s3:HeadObject permissions)"
	}

	// MinIO
	if strings.Contains(endpointLower, "minio") || strings.Contains(endpointLower, ":9000") {
		return " (MinIO: verify access key and secret key are correct)"
	}

	// DigitalOcean Spaces
	if strings.Contains(endpointLower, "digitaloceanspaces.com") {
		return " (DigitalOcean Spaces: verify endpoint region matches bucket region)"
	}

	// Backblaze B2
	if strings.Contains(endpointLower, "backblazeb2.com") {
		return " (Backblaze B2: check application key has read/write permissions)"
	}

	return ""
}
