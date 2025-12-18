package storage

import (
	"errors"
	"fmt"
	"log/slog"
)

var (
	// ErrTokenRequired is returned when a storage scheme requires a token but none was provided
	ErrTokenRequired = errors.New("storage token required")
)

// NewStorage creates a storage backend based on the URI scheme.
// Returns an appropriate Store implementation based on the URI scheme:
//   - file:// -> FileStorage
//   - oci:// -> OCIStorage (requires token)
//   - s3:// or s3+http:// -> S3Storage
func NewStorage(uri *StorageURI, token string, logger *slog.Logger) (Store, error) {
	switch uri.Scheme {
	case "file":
		return NewFileStorage(uri.Path, token, logger)

	case "oci":
		// Token is required for OCI storage
		if token == "" {
			return nil, fmt.Errorf("%w: OCI storage requires authentication token (--storage-token or COLA_REGISTRY_STORAGE_TOKEN)", ErrTokenRequired)
		}
		return NewOCIStorage(uri, token, logger)

	case "s3", "s3+http":
		// S3 storage (credentials optional for IAM role)
		return NewS3Storage(uri, token, logger)

	default:
		return nil, fmt.Errorf("unsupported storage scheme: %s", uri.Scheme)
	}
}
