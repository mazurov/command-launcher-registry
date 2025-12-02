package storage

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/criteo/command-launcher-registry/internal/models"
)

// OCIStorage implements Store interface using OCI registry as backend.
// It embeds BaseStorage for in-memory CRUD operations and provides
// OCI-based persistence via pushToOCI().
type OCIStorage struct {
	*BaseStorage       // Embedded for shared CRUD logic
	client       *OCIClient
	reference    string // OCI reference "registry/repo:latest"
}

// NewOCIStorage creates a new OCI-backed storage.
// The uri should be a parsed OCI StorageURI (oci://registry/repo).
// The token is used as a bearer token for OCI registry authentication.
func NewOCIStorage(uri *StorageURI, token string, logger *slog.Logger) (*OCIStorage, error) {
	if !uri.IsOCIScheme() {
		return nil, fmt.Errorf("expected OCI URI, got scheme: %s", uri.Scheme)
	}

	reference := uri.OCIReference()

	// Create OCI client
	client, err := NewOCIClient(reference, token, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create OCI client: %w", err)
	}

	s := &OCIStorage{
		BaseStorage: NewBaseStorage(logger),
		client:      client,
		reference:   reference,
	}

	// Load existing data from OCI or initialize empty storage
	if err := s.load(); err != nil {
		return nil, fmt.Errorf("failed to load data from OCI: %w", err)
	}

	return s, nil
}

// load retrieves registry data from OCI registry on startup.
// If the artifact doesn't exist, initializes empty storage and pushes it.
func (s *OCIStorage) load() error {
	ctx := context.Background()

	// Check if artifact exists
	exists, err := s.client.Exists(ctx)
	if err != nil {
		return fmt.Errorf("failed to check OCI artifact existence: %w", err)
	}

	if !exists {
		// Initialize empty storage and push to OCI
		s.logger.Info("OCI artifact does not exist, initializing empty storage",
			"reference", s.reference)

		// Push initial empty storage
		if err := s.persist(); err != nil {
			return fmt.Errorf("failed to initialize OCI storage: %w", err)
		}
		return nil
	}

	// Pull existing data
	data, err := s.client.Pull(ctx)
	if err != nil {
		return fmt.Errorf("failed to pull from OCI: %w", err)
	}

	// Parse JSON data
	if err := s.UnmarshalData(data); err != nil {
		return fmt.Errorf("failed to parse registry data (corrupted JSON): %w", err)
	}

	storageData := s.GetData()
	s.logger.Info("OCI storage loaded",
		"reference", s.reference,
		"registry_count", len(storageData.Registries))

	return nil
}

// persist pushes the complete registry data to OCI registry.
// NOTE: This is called while BaseStorage holds the lock,
// so we use marshalDataLocked() to avoid deadlock.
func (s *OCIStorage) persist() error {
	ctx := context.Background()

	data, err := s.marshalDataLocked()
	if err != nil {
		return fmt.Errorf("failed to marshal registry data: %w", err)
	}

	if err := s.client.Push(ctx, data); err != nil {
		return err // Already categorized by OCIClient
	}

	return nil
}

// CreateRegistry creates a new registry
func (s *OCIStorage) CreateRegistry(ctx context.Context, r *models.Registry) error {
	return s.BaseStorage.CreateRegistry(ctx, r, s.persist)
}

// GetRegistry retrieves a registry by name
func (s *OCIStorage) GetRegistry(ctx context.Context, name string) (*models.Registry, error) {
	return s.BaseStorage.GetRegistry(ctx, name)
}

// UpdateRegistry updates registry metadata
func (s *OCIStorage) UpdateRegistry(ctx context.Context, r *models.Registry) error {
	return s.BaseStorage.UpdateRegistry(ctx, r, s.persist)
}

// DeleteRegistry deletes a registry and all its packages (atomic)
func (s *OCIStorage) DeleteRegistry(ctx context.Context, name string) error {
	return s.BaseStorage.DeleteRegistry(ctx, name, s.persist)
}

// ListRegistries returns all registries
func (s *OCIStorage) ListRegistries(ctx context.Context) ([]*models.Registry, error) {
	return s.BaseStorage.ListRegistries(ctx)
}

// CreatePackage creates a new package in a registry
func (s *OCIStorage) CreatePackage(ctx context.Context, registryName string, p *models.Package) error {
	return s.BaseStorage.CreatePackage(ctx, registryName, p, s.persist)
}

// GetPackage retrieves a package from a registry
func (s *OCIStorage) GetPackage(ctx context.Context, registryName, packageName string) (*models.Package, error) {
	return s.BaseStorage.GetPackage(ctx, registryName, packageName)
}

// UpdatePackage updates package metadata (preserves versions)
func (s *OCIStorage) UpdatePackage(ctx context.Context, registryName string, p *models.Package) error {
	return s.BaseStorage.UpdatePackage(ctx, registryName, p, s.persist)
}

// DeletePackage deletes a package and all its versions (atomic)
func (s *OCIStorage) DeletePackage(ctx context.Context, registryName, packageName string) error {
	return s.BaseStorage.DeletePackage(ctx, registryName, packageName, s.persist)
}

// ListPackages returns all packages in a registry
func (s *OCIStorage) ListPackages(ctx context.Context, registryName string) ([]*models.Package, error) {
	return s.BaseStorage.ListPackages(ctx, registryName)
}

// CreateVersion creates a new version for a package
func (s *OCIStorage) CreateVersion(ctx context.Context, registryName, packageName string, v *models.Version) error {
	return s.BaseStorage.CreateVersion(ctx, registryName, packageName, v, s.persist)
}

// GetVersion retrieves a specific version
func (s *OCIStorage) GetVersion(ctx context.Context, registryName, packageName, version string) (*models.Version, error) {
	return s.BaseStorage.GetVersion(ctx, registryName, packageName, version)
}

// DeleteVersion deletes a specific version
func (s *OCIStorage) DeleteVersion(ctx context.Context, registryName, packageName, version string) error {
	return s.BaseStorage.DeleteVersion(ctx, registryName, packageName, version, s.persist)
}

// ListVersions returns all versions for a package
func (s *OCIStorage) ListVersions(ctx context.Context, registryName, packageName string) ([]*models.Version, error) {
	return s.BaseStorage.ListVersions(ctx, registryName, packageName)
}

// GetRegistryIndex generates the registry index (Command Launcher format)
func (s *OCIStorage) GetRegistryIndex(ctx context.Context, registryName string) ([]models.IndexEntry, error) {
	return s.BaseStorage.GetRegistryIndex(ctx, registryName)
}

// Close closes the storage (no-op for OCI storage)
func (s *OCIStorage) Close() error {
	return nil
}
