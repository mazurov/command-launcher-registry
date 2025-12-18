package storage

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/criteo/command-launcher-registry/internal/models"
)

// S3Storage implements Store interface using S3-compatible storage as backend.
// It embeds BaseStorage for in-memory CRUD operations and provides
// S3-based persistence via persist().
type S3Storage struct {
	*BaseStorage       // Embedded for shared CRUD logic
	client       *S3Client
	bucket       string
	key          string
}

// NewS3Storage creates a new S3-backed storage.
// The uri should be a parsed S3 StorageURI (s3://endpoint/bucket/path or s3+http://...).
// The token should be in format ACCESS_KEY:SECRET_KEY.
func NewS3Storage(uri *StorageURI, token string, logger *slog.Logger) (*S3Storage, error) {
	if !uri.IsS3Scheme() {
		return nil, fmt.Errorf("expected S3 URI, got scheme: %s", uri.Scheme)
	}

	// Extract S3 components from URI
	endpoint := uri.S3Endpoint()
	bucket := uri.S3Bucket()
	key := uri.S3Key()
	useSSL := uri.S3UseSSL()

	// Get region from URI query param or extract from endpoint
	region := uri.S3Region()
	if region == "" {
		region = ExtractRegionFromEndpoint(endpoint)
	}

	// Parse credentials from token
	accessKey, secretKey, err := ParseS3Token(token)
	if err != nil {
		return nil, fmt.Errorf("failed to parse S3 credentials: %w", err)
	}

	// Create S3 client
	client, err := NewS3Client(endpoint, bucket, key, accessKey, secretKey, useSSL, region, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 client: %w", err)
	}

	// Validate bucket exists
	ctx := context.Background()
	if err := client.ValidateBucket(ctx); err != nil {
		return nil, fmt.Errorf("S3 bucket validation failed: %w", err)
	}

	s := &S3Storage{
		BaseStorage: NewBaseStorage(logger),
		client:      client,
		bucket:      bucket,
		key:         key,
	}

	// Load existing data from S3 or initialize empty storage
	if err := s.load(); err != nil {
		return nil, fmt.Errorf("failed to load data from S3: %w", err)
	}

	return s, nil
}

// load retrieves registry data from S3 on startup.
// If the object doesn't exist, initializes empty storage and pushes it.
func (s *S3Storage) load() error {
	ctx := context.Background()

	// Check if object exists
	exists, err := s.client.Exists(ctx)
	if err != nil {
		return fmt.Errorf("failed to check S3 object existence: %w", err)
	}

	if !exists {
		// Initialize empty storage and push to S3
		s.logger.Info("S3 object does not exist, initializing empty storage",
			"bucket", s.bucket,
			"key", s.key)

		// Push initial empty storage
		if err := s.persist(); err != nil {
			return fmt.Errorf("failed to initialize S3 storage: %w", err)
		}
		return nil
	}

	// Download existing data
	data, err := s.client.Download(ctx)
	if err != nil {
		return fmt.Errorf("failed to download from S3: %w", err)
	}

	// Parse JSON data
	if err := s.UnmarshalData(data); err != nil {
		return fmt.Errorf("failed to parse registry data (corrupted JSON): %w", err)
	}

	storageData := s.GetData()
	s.logger.Info("S3 storage loaded",
		"bucket", s.bucket,
		"key", s.key,
		"registry_count", len(storageData.Registries))

	return nil
}

// persist uploads the complete registry data to S3.
// NOTE: This is called while BaseStorage holds the lock,
// so we use marshalDataLocked() to avoid deadlock.
func (s *S3Storage) persist() error {
	ctx := context.Background()

	data, err := s.marshalDataLocked()
	if err != nil {
		return fmt.Errorf("failed to marshal registry data: %w", err)
	}

	if err := s.client.Upload(ctx, data); err != nil {
		return err // Already categorized by S3Client
	}

	return nil
}

// CreateRegistry creates a new registry
func (s *S3Storage) CreateRegistry(ctx context.Context, r *models.Registry) error {
	return s.BaseStorage.CreateRegistry(ctx, r, s.persist)
}

// GetRegistry retrieves a registry by name
func (s *S3Storage) GetRegistry(ctx context.Context, name string) (*models.Registry, error) {
	return s.BaseStorage.GetRegistry(ctx, name)
}

// UpdateRegistry updates registry metadata
func (s *S3Storage) UpdateRegistry(ctx context.Context, r *models.Registry) error {
	return s.BaseStorage.UpdateRegistry(ctx, r, s.persist)
}

// DeleteRegistry deletes a registry and all its packages (atomic)
func (s *S3Storage) DeleteRegistry(ctx context.Context, name string) error {
	return s.BaseStorage.DeleteRegistry(ctx, name, s.persist)
}

// ListRegistries returns all registries
func (s *S3Storage) ListRegistries(ctx context.Context) ([]*models.Registry, error) {
	return s.BaseStorage.ListRegistries(ctx)
}

// CreatePackage creates a new package in a registry
func (s *S3Storage) CreatePackage(ctx context.Context, registryName string, p *models.Package) error {
	return s.BaseStorage.CreatePackage(ctx, registryName, p, s.persist)
}

// GetPackage retrieves a package from a registry
func (s *S3Storage) GetPackage(ctx context.Context, registryName, packageName string) (*models.Package, error) {
	return s.BaseStorage.GetPackage(ctx, registryName, packageName)
}

// UpdatePackage updates package metadata (preserves versions)
func (s *S3Storage) UpdatePackage(ctx context.Context, registryName string, p *models.Package) error {
	return s.BaseStorage.UpdatePackage(ctx, registryName, p, s.persist)
}

// DeletePackage deletes a package and all its versions (atomic)
func (s *S3Storage) DeletePackage(ctx context.Context, registryName, packageName string) error {
	return s.BaseStorage.DeletePackage(ctx, registryName, packageName, s.persist)
}

// ListPackages returns all packages in a registry
func (s *S3Storage) ListPackages(ctx context.Context, registryName string) ([]*models.Package, error) {
	return s.BaseStorage.ListPackages(ctx, registryName)
}

// CreateVersion creates a new version for a package
func (s *S3Storage) CreateVersion(ctx context.Context, registryName, packageName string, v *models.Version) error {
	return s.BaseStorage.CreateVersion(ctx, registryName, packageName, v, s.persist)
}

// GetVersion retrieves a specific version
func (s *S3Storage) GetVersion(ctx context.Context, registryName, packageName, version string) (*models.Version, error) {
	return s.BaseStorage.GetVersion(ctx, registryName, packageName, version)
}

// DeleteVersion deletes a specific version
func (s *S3Storage) DeleteVersion(ctx context.Context, registryName, packageName, version string) error {
	return s.BaseStorage.DeleteVersion(ctx, registryName, packageName, version, s.persist)
}

// ListVersions returns all versions for a package
func (s *S3Storage) ListVersions(ctx context.Context, registryName, packageName string) ([]*models.Version, error) {
	return s.BaseStorage.ListVersions(ctx, registryName, packageName)
}

// GetRegistryIndex generates the registry index (Command Launcher format)
func (s *S3Storage) GetRegistryIndex(ctx context.Context, registryName string) ([]models.IndexEntry, error) {
	return s.BaseStorage.GetRegistryIndex(ctx, registryName)
}

// Close closes the storage (no-op for S3 storage)
func (s *S3Storage) Close() error {
	return nil
}
