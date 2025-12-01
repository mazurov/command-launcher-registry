package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/criteo/command-launcher-registry/internal/models"
)

// FileStorage implements Store interface using file-based storage
type FileStorage struct {
	filePath string
	mu       sync.RWMutex
	data     *models.Storage
	logger   *slog.Logger
}

// NewFileStorage creates a new file-based storage
// The token parameter is accepted but ignored for file storage (for interface compatibility)
func NewFileStorage(filePath string, token string, logger *slog.Logger) (*FileStorage, error) {
	// Log warning if token is provided (file storage doesn't use it)
	if token != "" {
		logger.Warn("Storage token provided but file storage does not use authentication",
			"file_path", filePath)
	}

	fs := &FileStorage{
		filePath: filePath,
		logger:   logger,
	}

	// Load existing data or create new storage
	if err := fs.load(); err != nil {
		return nil, fmt.Errorf("failed to load storage: %w", err)
	}

	return fs, nil
}

// load reads storage from file or creates empty storage
func (fs *FileStorage) load() error {
	// Check if file exists
	if _, err := os.Stat(fs.filePath); os.IsNotExist(err) {
		// Create empty storage
		fs.data = models.NewStorage()
		fs.logger.Info("Storage file not found, creating empty storage",
			"file_path", fs.filePath)

		// Create directory if needed
		dir := filepath.Dir(fs.filePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create storage directory: %w", err)
		}

		// Write empty storage to file
		if err := fs.saveToFile(); err != nil {
			return fmt.Errorf("failed to create storage file: %w", err)
		}

		return nil
	}

	// Read existing file
	fileData, err := os.ReadFile(fs.filePath)
	if err != nil {
		return fmt.Errorf("failed to read storage file: %w", err)
	}

	// Parse JSON
	var data models.Storage
	if err := json.Unmarshal(fileData, &data); err != nil {
		return fmt.Errorf("failed to parse storage file (invalid JSON syntax): %w", err)
	}

	// Initialize maps if nil
	if data.Registries == nil {
		data.Registries = make(map[string]*models.Registry)
	}

	fs.data = &data
	fs.logger.Info("Storage file loaded",
		"file_path", fs.filePath,
		"registry_count", len(fs.data.Registries))

	return nil
}

// saveToFile writes data to file atomically (temp file + rename)
func (fs *FileStorage) saveToFile() error {
	// Marshal to JSON
	jsonData, err := json.MarshalIndent(fs.data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal storage: %w", err)
	}

	// Create temp file in same directory
	dir := filepath.Dir(fs.filePath)
	tempFile, err := os.CreateTemp(dir, ".registry-*.json.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tempPath := tempFile.Name()

	// Ensure temp file cleanup on error
	defer func() {
		if tempFile != nil {
			tempFile.Close()
			os.Remove(tempPath)
		}
	}()

	// Write to temp file
	if _, err := tempFile.Write(jsonData); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Sync to disk
	if err := tempFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync temp file: %w", err)
	}

	// Close temp file
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}
	tempFile = nil // Prevent deferred cleanup

	// Atomic rename
	if err := os.Rename(tempPath, fs.filePath); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	// Check file size and warn if > 50MB
	if info, err := os.Stat(fs.filePath); err == nil {
		sizeMB := float64(info.Size()) / (1024 * 1024)
		if sizeMB > 50 {
			fs.logger.Warn("Storage file size exceeds recommended threshold",
				"file_path", fs.filePath,
				"current_size_mb", sizeMB,
				"threshold_mb", 50,
				"max_size_mb", 100,
				"registries_count", len(fs.data.Registries),
			)
		}
	}

	return nil
}

// CreateRegistry creates a new registry
func (fs *FileStorage) CreateRegistry(ctx context.Context, r *models.Registry) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Check if already exists
	if _, exists := fs.data.Registries[r.Name]; exists {
		return ErrAlreadyExists
	}

	// Add to storage
	fs.data.Registries[r.Name] = r

	// Persist to disk
	if err := fs.saveToFile(); err != nil {
		// Rollback in-memory change
		delete(fs.data.Registries, r.Name)
		fs.logger.Error("Storage write failed",
			"operation", "create_registry",
			"registry", r.Name,
			"error", err)
		return ErrStorageUnavailable
	}

	fs.logger.Info("Registry created", "registry", r.Name)
	return nil
}

// GetRegistry retrieves a registry by name
func (fs *FileStorage) GetRegistry(ctx context.Context, name string) (*models.Registry, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	registry, exists := fs.data.Registries[name]
	if !exists {
		return nil, ErrNotFound
	}

	return registry, nil
}

// UpdateRegistry updates registry metadata
func (fs *FileStorage) UpdateRegistry(ctx context.Context, r *models.Registry) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Check if exists
	existing, exists := fs.data.Registries[r.Name]
	if !exists {
		return ErrNotFound
	}

	// Preserve packages
	r.Packages = existing.Packages

	// Update in storage
	fs.data.Registries[r.Name] = r

	// Persist to disk
	if err := fs.saveToFile(); err != nil {
		// Rollback
		fs.data.Registries[r.Name] = existing
		fs.logger.Error("Storage write failed",
			"operation", "update_registry",
			"registry", r.Name,
			"error", err)
		return ErrStorageUnavailable
	}

	fs.logger.Info("Registry updated", "registry", r.Name)
	return nil
}

// DeleteRegistry deletes a registry and all its packages (atomic)
func (fs *FileStorage) DeleteRegistry(ctx context.Context, name string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Check if exists
	registry, exists := fs.data.Registries[name]
	if !exists {
		return ErrNotFound
	}

	// Delete from storage (in-memory)
	delete(fs.data.Registries, name)

	// Persist to disk
	if err := fs.saveToFile(); err != nil {
		// Rollback
		fs.data.Registries[name] = registry
		fs.logger.Error("Storage write failed",
			"operation", "delete_registry",
			"registry", name,
			"error", err)
		return ErrStorageUnavailable
	}

	fs.logger.Info("Registry deleted",
		"registry", name,
		"packages_deleted", len(registry.Packages))
	return nil
}

// ListRegistries returns all registries
func (fs *FileStorage) ListRegistries(ctx context.Context) ([]*models.Registry, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	registries := make([]*models.Registry, 0, len(fs.data.Registries))
	for _, r := range fs.data.Registries {
		registries = append(registries, r)
	}

	return registries, nil
}

// CreatePackage creates a new package in a registry
func (fs *FileStorage) CreatePackage(ctx context.Context, registryName string, p *models.Package) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Get registry
	registry, exists := fs.data.Registries[registryName]
	if !exists {
		return ErrNotFound
	}

	// Check if package already exists
	if _, exists := registry.Packages[p.Name]; exists {
		return ErrAlreadyExists
	}

	// Add package
	registry.Packages[p.Name] = p

	// Persist to disk
	if err := fs.saveToFile(); err != nil {
		// Rollback
		delete(registry.Packages, p.Name)
		fs.logger.Error("Storage write failed",
			"operation", "create_package",
			"registry", registryName,
			"package", p.Name,
			"error", err)
		return ErrStorageUnavailable
	}

	fs.logger.Info("Package created",
		"registry", registryName,
		"package", p.Name)
	return nil
}

// GetPackage retrieves a package from a registry
func (fs *FileStorage) GetPackage(ctx context.Context, registryName, packageName string) (*models.Package, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	registry, exists := fs.data.Registries[registryName]
	if !exists {
		return nil, ErrNotFound
	}

	pkg, exists := registry.Packages[packageName]
	if !exists {
		return nil, ErrNotFound
	}

	return pkg, nil
}

// UpdatePackage updates package metadata (preserves versions)
func (fs *FileStorage) UpdatePackage(ctx context.Context, registryName string, p *models.Package) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Get registry
	registry, exists := fs.data.Registries[registryName]
	if !exists {
		return ErrNotFound
	}

	// Check if package exists
	if _, exists := registry.Packages[p.Name]; !exists {
		return ErrNotFound
	}

	// Store old package for rollback
	oldPackage := registry.Packages[p.Name]

	// Update package
	registry.Packages[p.Name] = p

	// Persist to disk
	if err := fs.saveToFile(); err != nil {
		// Rollback
		registry.Packages[p.Name] = oldPackage
		fs.logger.Error("Storage write failed",
			"operation", "update_package",
			"registry", registryName,
			"package", p.Name,
			"error", err)
		return ErrStorageUnavailable
	}

	fs.logger.Info("Package updated",
		"registry", registryName,
		"package", p.Name)

	return nil
}

// DeletePackage deletes a package and all its versions (atomic)
func (fs *FileStorage) DeletePackage(ctx context.Context, registryName, packageName string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Get registry
	registry, exists := fs.data.Registries[registryName]
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

	// Persist to disk
	if err := fs.saveToFile(); err != nil {
		// Rollback
		registry.Packages[packageName] = pkg
		fs.logger.Error("Storage write failed",
			"operation", "delete_package",
			"registry", registryName,
			"package", packageName,
			"error", err)
		return ErrStorageUnavailable
	}

	fs.logger.Info("Package deleted",
		"registry", registryName,
		"package", packageName,
		"versions_deleted", len(pkg.Versions))
	return nil
}

// ListPackages returns all packages in a registry
func (fs *FileStorage) ListPackages(ctx context.Context, registryName string) ([]*models.Package, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	registry, exists := fs.data.Registries[registryName]
	if !exists {
		return nil, ErrNotFound
	}

	packages := make([]*models.Package, 0, len(registry.Packages))
	for _, p := range registry.Packages {
		packages = append(packages, p)
	}

	return packages, nil
}

// CreateVersion creates a new version for a package
func (fs *FileStorage) CreateVersion(ctx context.Context, registryName, packageName string, v *models.Version) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Get registry
	registry, exists := fs.data.Registries[registryName]
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

	// Persist to disk
	if err := fs.saveToFile(); err != nil {
		// Rollback
		delete(pkg.Versions, v.Version)
		fs.logger.Error("Storage write failed",
			"operation", "create_version",
			"registry", registryName,
			"package", packageName,
			"version", v.Version,
			"error", err)
		return ErrStorageUnavailable
	}

	fs.logger.Info("Version created",
		"registry", registryName,
		"package", packageName,
		"version", v.Version)
	return nil
}

// GetVersion retrieves a specific version
func (fs *FileStorage) GetVersion(ctx context.Context, registryName, packageName, version string) (*models.Version, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	registry, exists := fs.data.Registries[registryName]
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

// DeleteVersion deletes a specific version
func (fs *FileStorage) DeleteVersion(ctx context.Context, registryName, packageName, version string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Get registry
	registry, exists := fs.data.Registries[registryName]
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

	// Persist to disk
	if err := fs.saveToFile(); err != nil {
		// Rollback
		pkg.Versions[version] = ver
		fs.logger.Error("Storage write failed",
			"operation", "delete_version",
			"registry", registryName,
			"package", packageName,
			"version", version,
			"error", err)
		return ErrStorageUnavailable
	}

	fs.logger.Info("Version deleted",
		"registry", registryName,
		"package", packageName,
		"version", version)
	return nil
}

// ListVersions returns all versions for a package
func (fs *FileStorage) ListVersions(ctx context.Context, registryName, packageName string) ([]*models.Version, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	registry, exists := fs.data.Registries[registryName]
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
func (fs *FileStorage) GetRegistryIndex(ctx context.Context, registryName string) ([]models.IndexEntry, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	registry, exists := fs.data.Registries[registryName]
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

// Close closes the storage (no-op for file storage)
func (fs *FileStorage) Close() error {
	return nil
}
