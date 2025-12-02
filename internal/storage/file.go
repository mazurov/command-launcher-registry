package storage

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/criteo/command-launcher-registry/internal/models"
)

// FileStorage implements Store interface using file-based storage.
// It embeds BaseStorage for in-memory CRUD operations and provides
// file-based persistence via saveToFile().
type FileStorage struct {
	*BaseStorage         // Embedded for shared CRUD logic
	filePath     string  // Path to storage file
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
		BaseStorage: NewBaseStorage(logger),
		filePath:    filePath,
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
		// Create empty storage (already initialized in NewBaseStorage)
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

	// Parse JSON using BaseStorage's unmarshal
	if err := fs.UnmarshalData(fileData); err != nil {
		return fmt.Errorf("failed to parse storage file (invalid JSON syntax): %w", err)
	}

	data := fs.GetData()
	fs.logger.Info("Storage file loaded",
		"file_path", fs.filePath,
		"registry_count", len(data.Registries))

	return nil
}

// saveToFile writes data to file atomically (temp file + rename)
// NOTE: This is called from persist() while BaseStorage holds the lock,
// so we use marshalDataLocked() to avoid deadlock.
func (fs *FileStorage) saveToFile() error {
	// Marshal to JSON using lock-free version (caller holds lock)
	jsonData, err := fs.marshalDataLocked()
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
			data := fs.getDataLocked() // Use lock-free version (caller holds lock)
			fs.logger.Warn("Storage file size exceeds recommended threshold",
				"file_path", fs.filePath,
				"current_size_mb", sizeMB,
				"threshold_mb", 50,
				"max_size_mb", 100,
				"registries_count", len(data.Registries),
			)
		}
	}

	return nil
}

// persist is the callback passed to BaseStorage methods
func (fs *FileStorage) persist() error {
	return fs.saveToFile()
}

// CreateRegistry creates a new registry
func (fs *FileStorage) CreateRegistry(ctx context.Context, r *models.Registry) error {
	return fs.BaseStorage.CreateRegistry(ctx, r, fs.persist)
}

// GetRegistry retrieves a registry by name
func (fs *FileStorage) GetRegistry(ctx context.Context, name string) (*models.Registry, error) {
	return fs.BaseStorage.GetRegistry(ctx, name)
}

// UpdateRegistry updates registry metadata
func (fs *FileStorage) UpdateRegistry(ctx context.Context, r *models.Registry) error {
	return fs.BaseStorage.UpdateRegistry(ctx, r, fs.persist)
}

// DeleteRegistry deletes a registry and all its packages (atomic)
func (fs *FileStorage) DeleteRegistry(ctx context.Context, name string) error {
	return fs.BaseStorage.DeleteRegistry(ctx, name, fs.persist)
}

// ListRegistries returns all registries
func (fs *FileStorage) ListRegistries(ctx context.Context) ([]*models.Registry, error) {
	return fs.BaseStorage.ListRegistries(ctx)
}

// CreatePackage creates a new package in a registry
func (fs *FileStorage) CreatePackage(ctx context.Context, registryName string, p *models.Package) error {
	return fs.BaseStorage.CreatePackage(ctx, registryName, p, fs.persist)
}

// GetPackage retrieves a package from a registry
func (fs *FileStorage) GetPackage(ctx context.Context, registryName, packageName string) (*models.Package, error) {
	return fs.BaseStorage.GetPackage(ctx, registryName, packageName)
}

// UpdatePackage updates package metadata (preserves versions)
func (fs *FileStorage) UpdatePackage(ctx context.Context, registryName string, p *models.Package) error {
	return fs.BaseStorage.UpdatePackage(ctx, registryName, p, fs.persist)
}

// DeletePackage deletes a package and all its versions (atomic)
func (fs *FileStorage) DeletePackage(ctx context.Context, registryName, packageName string) error {
	return fs.BaseStorage.DeletePackage(ctx, registryName, packageName, fs.persist)
}

// ListPackages returns all packages in a registry
func (fs *FileStorage) ListPackages(ctx context.Context, registryName string) ([]*models.Package, error) {
	return fs.BaseStorage.ListPackages(ctx, registryName)
}

// CreateVersion creates a new version for a package
func (fs *FileStorage) CreateVersion(ctx context.Context, registryName, packageName string, v *models.Version) error {
	return fs.BaseStorage.CreateVersion(ctx, registryName, packageName, v, fs.persist)
}

// GetVersion retrieves a specific version
func (fs *FileStorage) GetVersion(ctx context.Context, registryName, packageName, version string) (*models.Version, error) {
	return fs.BaseStorage.GetVersion(ctx, registryName, packageName, version)
}

// DeleteVersion deletes a specific version
func (fs *FileStorage) DeleteVersion(ctx context.Context, registryName, packageName, version string) error {
	return fs.BaseStorage.DeleteVersion(ctx, registryName, packageName, version, fs.persist)
}

// ListVersions returns all versions for a package
func (fs *FileStorage) ListVersions(ctx context.Context, registryName, packageName string) ([]*models.Version, error) {
	return fs.BaseStorage.ListVersions(ctx, registryName, packageName)
}

// GetRegistryIndex generates the registry index (Command Launcher format)
func (fs *FileStorage) GetRegistryIndex(ctx context.Context, registryName string) ([]models.IndexEntry, error) {
	return fs.BaseStorage.GetRegistryIndex(ctx, registryName)
}

// Close closes the storage (no-op for file storage)
func (fs *FileStorage) Close() error {
	return nil
}
