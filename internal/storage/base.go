package storage

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/criteo/command-launcher-registry/internal/models"
)

// BaseStorage provides shared in-memory CRUD operations for all storage backends.
// It handles locking, validation, and data manipulation. Concrete backends (FileStorage,
// OCIStorage) embed this and provide their own persistence mechanisms.
type BaseStorage struct {
	mu     sync.RWMutex
	data   *models.Storage
	logger *slog.Logger
}

// NewBaseStorage creates a new BaseStorage with empty data
func NewBaseStorage(logger *slog.Logger) *BaseStorage {
	return &BaseStorage{
		data:   models.NewStorage(),
		logger: logger,
	}
}

// SetData sets the in-memory data (used by backends after loading)
func (b *BaseStorage) SetData(data *models.Storage) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.data = data
}

// GetData returns a copy of the current data (used by backends for persistence)
func (b *BaseStorage) GetData() *models.Storage {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.data
}

// MarshalData serializes the storage data to JSON.
// NOTE: Caller must NOT hold the lock - this method acquires its own lock.
// For use within locked contexts, use marshalDataLocked instead.
func (b *BaseStorage) MarshalData() ([]byte, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return json.MarshalIndent(b.data, "", "  ")
}

// marshalDataLocked serializes data without acquiring lock.
// Caller MUST hold at least a read lock.
func (b *BaseStorage) marshalDataLocked() ([]byte, error) {
	return json.MarshalIndent(b.data, "", "  ")
}

// getDataLocked returns the data without acquiring lock.
// Caller MUST hold at least a read lock.
func (b *BaseStorage) getDataLocked() *models.Storage {
	return b.data
}

// UnmarshalData deserializes JSON data into storage
func (b *BaseStorage) UnmarshalData(jsonData []byte) error {
	var data models.Storage
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return err
	}
	// Initialize maps if nil
	if data.Registries == nil {
		data.Registries = make(map[string]*models.Registry)
	}
	b.mu.Lock()
	b.data = &data
	b.mu.Unlock()
	return nil
}

// PersistFunc is a callback function that backends implement for persistence
type PersistFunc func() error

// CreateRegistry creates a new registry in memory.
// The persist callback is called after the in-memory operation succeeds.
// If persist fails, the in-memory change is rolled back.
func (b *BaseStorage) CreateRegistry(ctx context.Context, r *models.Registry, persist PersistFunc) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Check if already exists
	if _, exists := b.data.Registries[r.Name]; exists {
		return ErrAlreadyExists
	}

	// Add to storage
	b.data.Registries[r.Name] = r

	// Persist
	if persist != nil {
		if err := persist(); err != nil {
			// Rollback in-memory change
			delete(b.data.Registries, r.Name)
			b.logger.Error("Storage write failed",
				"operation", "create_registry",
				"registry", r.Name,
				"error", err)
			return ErrStorageUnavailable
		}
	}

	b.logger.Info("Registry created", "registry", r.Name)
	return nil
}

// GetRegistry retrieves a registry by name
func (b *BaseStorage) GetRegistry(ctx context.Context, name string) (*models.Registry, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	registry, exists := b.data.Registries[name]
	if !exists {
		return nil, ErrNotFound
	}

	return registry, nil
}

// UpdateRegistry updates registry metadata.
// The persist callback is called after the in-memory operation succeeds.
func (b *BaseStorage) UpdateRegistry(ctx context.Context, r *models.Registry, persist PersistFunc) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Check if exists
	existing, exists := b.data.Registries[r.Name]
	if !exists {
		return ErrNotFound
	}

	// Preserve packages
	r.Packages = existing.Packages

	// Update in storage
	b.data.Registries[r.Name] = r

	// Persist
	if persist != nil {
		if err := persist(); err != nil {
			// Rollback
			b.data.Registries[r.Name] = existing
			b.logger.Error("Storage write failed",
				"operation", "update_registry",
				"registry", r.Name,
				"error", err)
			return ErrStorageUnavailable
		}
	}

	b.logger.Info("Registry updated", "registry", r.Name)
	return nil
}

// DeleteRegistry deletes a registry and all its packages.
// The persist callback is called after the in-memory operation succeeds.
func (b *BaseStorage) DeleteRegistry(ctx context.Context, name string, persist PersistFunc) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Check if exists
	registry, exists := b.data.Registries[name]
	if !exists {
		return ErrNotFound
	}

	// Delete from storage (in-memory)
	delete(b.data.Registries, name)

	// Persist
	if persist != nil {
		if err := persist(); err != nil {
			// Rollback
			b.data.Registries[name] = registry
			b.logger.Error("Storage write failed",
				"operation", "delete_registry",
				"registry", name,
				"error", err)
			return ErrStorageUnavailable
		}
	}

	b.logger.Info("Registry deleted",
		"registry", name,
		"packages_deleted", len(registry.Packages))
	return nil
}

// ListRegistries returns all registries
func (b *BaseStorage) ListRegistries(ctx context.Context) ([]*models.Registry, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	registries := make([]*models.Registry, 0, len(b.data.Registries))
	for _, r := range b.data.Registries {
		registries = append(registries, r)
	}

	return registries, nil
}

// CreatePackage creates a new package in a registry.
// The persist callback is called after the in-memory operation succeeds.
func (b *BaseStorage) CreatePackage(ctx context.Context, registryName string, p *models.Package, persist PersistFunc) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Get registry
	registry, exists := b.data.Registries[registryName]
	if !exists {
		return ErrNotFound
	}

	// Check if package already exists
	if _, exists := registry.Packages[p.Name]; exists {
		return ErrAlreadyExists
	}

	// Add package
	registry.Packages[p.Name] = p

	// Persist
	if persist != nil {
		if err := persist(); err != nil {
			// Rollback
			delete(registry.Packages, p.Name)
			b.logger.Error("Storage write failed",
				"operation", "create_package",
				"registry", registryName,
				"package", p.Name,
				"error", err)
			return ErrStorageUnavailable
		}
	}

	b.logger.Info("Package created",
		"registry", registryName,
		"package", p.Name)
	return nil
}

// GetPackage retrieves a package from a registry
func (b *BaseStorage) GetPackage(ctx context.Context, registryName, packageName string) (*models.Package, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	registry, exists := b.data.Registries[registryName]
	if !exists {
		return nil, ErrNotFound
	}

	pkg, exists := registry.Packages[packageName]
	if !exists {
		return nil, ErrNotFound
	}

	return pkg, nil
}

// UpdatePackage updates package metadata (preserves versions).
// The persist callback is called after the in-memory operation succeeds.
func (b *BaseStorage) UpdatePackage(ctx context.Context, registryName string, p *models.Package, persist PersistFunc) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Get registry
	registry, exists := b.data.Registries[registryName]
	if !exists {
		return ErrNotFound
	}

	// Check if package exists
	oldPackage, exists := registry.Packages[p.Name]
	if !exists {
		return ErrNotFound
	}

	// Update package
	registry.Packages[p.Name] = p

	// Persist
	if persist != nil {
		if err := persist(); err != nil {
			// Rollback
			registry.Packages[p.Name] = oldPackage
			b.logger.Error("Storage write failed",
				"operation", "update_package",
				"registry", registryName,
				"package", p.Name,
				"error", err)
			return ErrStorageUnavailable
		}
	}

	b.logger.Info("Package updated",
		"registry", registryName,
		"package", p.Name)

	return nil
}

// DeletePackage deletes a package and all its versions.
// The persist callback is called after the in-memory operation succeeds.
func (b *BaseStorage) DeletePackage(ctx context.Context, registryName, packageName string, persist PersistFunc) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Get registry
	registry, exists := b.data.Registries[registryName]
	if !exists {
		return ErrNotFound
	}

	// Get package
	pkg, exists := registry.Packages[packageName]
	if !exists {
		return ErrNotFound
	}

	// Delete package
	delete(registry.Packages, packageName)

	// Persist
	if persist != nil {
		if err := persist(); err != nil {
			// Rollback
			registry.Packages[packageName] = pkg
			b.logger.Error("Storage write failed",
				"operation", "delete_package",
				"registry", registryName,
				"package", packageName,
				"error", err)
			return ErrStorageUnavailable
		}
	}

	b.logger.Info("Package deleted",
		"registry", registryName,
		"package", packageName,
		"versions_deleted", len(pkg.Versions))
	return nil
}

// ListPackages returns all packages in a registry
func (b *BaseStorage) ListPackages(ctx context.Context, registryName string) ([]*models.Package, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	registry, exists := b.data.Registries[registryName]
	if !exists {
		return nil, ErrNotFound
	}

	packages := make([]*models.Package, 0, len(registry.Packages))
	for _, p := range registry.Packages {
		packages = append(packages, p)
	}

	return packages, nil
}

// CreateVersion creates a new version for a package.
// Enforces immutability and partition overlap validation.
// The persist callback is called after the in-memory operation succeeds.
func (b *BaseStorage) CreateVersion(ctx context.Context, registryName, packageName string, v *models.Version, persist PersistFunc) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Get registry
	registry, exists := b.data.Registries[registryName]
	if !exists {
		return ErrNotFound
	}

	// Get package
	pkg, exists := registry.Packages[packageName]
	if !exists {
		return ErrNotFound
	}

	// Check if version already exists (immutability)
	if _, exists := pkg.Versions[v.Version]; exists {
		return ErrImmutabilityViolation
	}

	// Check for partition overlaps with existing versions
	for _, existingVersion := range pkg.Versions {
		if models.CheckPartitionOverlap(
			v.StartPartition, v.EndPartition,
			existingVersion.StartPartition, existingVersion.EndPartition,
		) {
			return ErrPartitionOverlap
		}
	}

	// Add version
	pkg.Versions[v.Version] = v

	// Persist
	if persist != nil {
		if err := persist(); err != nil {
			// Rollback
			delete(pkg.Versions, v.Version)
			b.logger.Error("Storage write failed",
				"operation", "create_version",
				"registry", registryName,
				"package", packageName,
				"version", v.Version,
				"error", err)
			return ErrStorageUnavailable
		}
	}

	b.logger.Info("Version created",
		"registry", registryName,
		"package", packageName,
		"version", v.Version)
	return nil
}

// GetVersion retrieves a specific version
func (b *BaseStorage) GetVersion(ctx context.Context, registryName, packageName, version string) (*models.Version, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	registry, exists := b.data.Registries[registryName]
	if !exists {
		return nil, ErrNotFound
	}

	pkg, exists := registry.Packages[packageName]
	if !exists {
		return nil, ErrNotFound
	}

	ver, exists := pkg.Versions[version]
	if !exists {
		return nil, ErrNotFound
	}

	return ver, nil
}

// DeleteVersion deletes a specific version.
// The persist callback is called after the in-memory operation succeeds.
func (b *BaseStorage) DeleteVersion(ctx context.Context, registryName, packageName, version string, persist PersistFunc) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Get registry
	registry, exists := b.data.Registries[registryName]
	if !exists {
		return ErrNotFound
	}

	// Get package
	pkg, exists := registry.Packages[packageName]
	if !exists {
		return ErrNotFound
	}

	// Get version
	ver, exists := pkg.Versions[version]
	if !exists {
		return ErrNotFound
	}

	// Delete version
	delete(pkg.Versions, version)

	// Persist
	if persist != nil {
		if err := persist(); err != nil {
			// Rollback
			pkg.Versions[version] = ver
			b.logger.Error("Storage write failed",
				"operation", "delete_version",
				"registry", registryName,
				"package", packageName,
				"version", version,
				"error", err)
			return ErrStorageUnavailable
		}
	}

	b.logger.Info("Version deleted",
		"registry", registryName,
		"package", packageName,
		"version", version)
	return nil
}

// ListVersions returns all versions for a package
func (b *BaseStorage) ListVersions(ctx context.Context, registryName, packageName string) ([]*models.Version, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	registry, exists := b.data.Registries[registryName]
	if !exists {
		return nil, ErrNotFound
	}

	pkg, exists := registry.Packages[packageName]
	if !exists {
		return nil, ErrNotFound
	}

	versions := make([]*models.Version, 0, len(pkg.Versions))
	for _, v := range pkg.Versions {
		versions = append(versions, v)
	}

	return versions, nil
}

// GetRegistryIndex generates the registry index (Command Launcher format)
func (b *BaseStorage) GetRegistryIndex(ctx context.Context, registryName string) ([]models.IndexEntry, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	registry, exists := b.data.Registries[registryName]
	if !exists {
		return nil, ErrNotFound
	}

	// Flatten all package versions into index entries
	var entries []models.IndexEntry
	for _, pkg := range registry.Packages {
		for _, ver := range pkg.Versions {
			entries = append(entries, ver.ToIndexEntry())
		}
	}

	return entries, nil
}
