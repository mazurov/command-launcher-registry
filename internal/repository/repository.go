package repository

import (
	"errors"
	"fmt"

	"github.com/mazurov/command-launcher-registry/internal/models"
	"gorm.io/gorm"
)

var (
	ErrRegistryNotFound = errors.New("registry not found")
	ErrPackageNotFound  = errors.New("package not found")
	ErrVersionNotFound  = errors.New("version not found")
	ErrAlreadyExists    = errors.New("resource already exists")
)

// Repository provides data access methods
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new repository instance
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// ==================== Registry Operations ====================

// CreateRegistry creates a new registry
func (r *Repository) CreateRegistry(registry *models.Registry) error {
	result := r.db.Create(registry)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return fmt.Errorf("%w: registry '%s' already exists", ErrAlreadyExists, registry.Name)
		}
		return result.Error
	}
	return nil
}

// GetRegistry retrieves a registry by name
func (r *Repository) GetRegistry(name string) (*models.Registry, error) {
	var registry models.Registry
	result := r.db.Where("name = ?", name).First(&registry)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRegistryNotFound
		}
		return nil, result.Error
	}
	return &registry, nil
}

// GetRegistryWithPackages retrieves a registry with all its packages and versions
func (r *Repository) GetRegistryWithPackages(name string) (*models.Registry, error) {
	var registry models.Registry
	result := r.db.Preload("Packages.Versions").Where("name = ?", name).First(&registry)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRegistryNotFound
		}
		return nil, result.Error
	}
	return &registry, nil
}

// ListRegistries returns all registries
func (r *Repository) ListRegistries() ([]models.Registry, error) {
	var registries []models.Registry
	result := r.db.Find(&registries)
	if result.Error != nil {
		return nil, result.Error
	}
	return registries, nil
}

// UpdateRegistry updates an existing registry
func (r *Repository) UpdateRegistry(name string, updates map[string]interface{}) error {
	result := r.db.Model(&models.Registry{}).Where("name = ?", name).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrRegistryNotFound
	}
	return nil
}

// DeleteRegistry deletes a registry (soft delete)
func (r *Repository) DeleteRegistry(name string) error {
	result := r.db.Where("name = ?", name).Delete(&models.Registry{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrRegistryNotFound
	}
	return nil
}

// ==================== Package Operations ====================

// CreatePackage creates a new package within a registry
func (r *Repository) CreatePackage(registryName string, pkg *models.Package) error {
	// First get the registry
	registry, err := r.GetRegistry(registryName)
	if err != nil {
		return err
	}

	pkg.RegistryID = registry.ID

	// Check if package already exists in this registry
	var existing models.Package
	result := r.db.Where("registry_id = ? AND name = ?", registry.ID, pkg.Name).First(&existing)
	if result.Error == nil {
		return fmt.Errorf("%w: package '%s' already exists in registry '%s'", ErrAlreadyExists, pkg.Name, registryName)
	}

	result = r.db.Create(pkg)
	return result.Error
}

// GetPackage retrieves a package by registry and package name
func (r *Repository) GetPackage(registryName, packageName string) (*models.Package, error) {
	registry, err := r.GetRegistry(registryName)
	if err != nil {
		return nil, err
	}

	var pkg models.Package
	result := r.db.Where("registry_id = ? AND name = ?", registry.ID, packageName).First(&pkg)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrPackageNotFound
		}
		return nil, result.Error
	}
	return &pkg, nil
}

// GetPackageWithVersions retrieves a package with all its versions
func (r *Repository) GetPackageWithVersions(registryName, packageName string) (*models.Package, error) {
	registry, err := r.GetRegistry(registryName)
	if err != nil {
		return nil, err
	}

	var pkg models.Package
	result := r.db.Preload("Versions").Where("registry_id = ? AND name = ?", registry.ID, packageName).First(&pkg)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrPackageNotFound
		}
		return nil, result.Error
	}
	return &pkg, nil
}

// ListPackages returns all packages in a registry
func (r *Repository) ListPackages(registryName string) ([]models.Package, error) {
	registry, err := r.GetRegistry(registryName)
	if err != nil {
		return nil, err
	}

	var packages []models.Package
	result := r.db.Where("registry_id = ?", registry.ID).Find(&packages)
	if result.Error != nil {
		return nil, result.Error
	}
	return packages, nil
}

// UpdatePackage updates an existing package
func (r *Repository) UpdatePackage(registryName, packageName string, updates map[string]interface{}) error {
	pkg, err := r.GetPackage(registryName, packageName)
	if err != nil {
		return err
	}

	result := r.db.Model(&models.Package{}).Where("id = ?", pkg.ID).Updates(updates)
	return result.Error
}

// DeletePackage deletes a package (soft delete)
func (r *Repository) DeletePackage(registryName, packageName string) error {
	pkg, err := r.GetPackage(registryName, packageName)
	if err != nil {
		return err
	}

	result := r.db.Delete(&models.Package{}, pkg.ID)
	return result.Error
}

// ==================== Version Operations ====================

// CreateVersion creates a new version for a package
func (r *Repository) CreateVersion(registryName, packageName string, version *models.Version) error {
	pkg, err := r.GetPackage(registryName, packageName)
	if err != nil {
		return err
	}

	version.PackageID = pkg.ID

	// Check if version already exists
	var existing models.Version
	result := r.db.Where("package_id = ? AND version = ?", pkg.ID, version.Version).First(&existing)
	if result.Error == nil {
		return fmt.Errorf("%w: version '%s' already exists for package '%s'", ErrAlreadyExists, version.Version, packageName)
	}

	result = r.db.Create(version)
	return result.Error
}

// GetVersion retrieves a specific version
func (r *Repository) GetVersion(registryName, packageName, versionStr string) (*models.Version, error) {
	pkg, err := r.GetPackage(registryName, packageName)
	if err != nil {
		return nil, err
	}

	var version models.Version
	result := r.db.Where("package_id = ? AND version = ?", pkg.ID, versionStr).First(&version)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrVersionNotFound
		}
		return nil, result.Error
	}
	return &version, nil
}

// ListVersions returns all versions of a package
func (r *Repository) ListVersions(registryName, packageName string) ([]models.Version, error) {
	pkg, err := r.GetPackage(registryName, packageName)
	if err != nil {
		return nil, err
	}

	var versions []models.Version
	result := r.db.Where("package_id = ?", pkg.ID).Order("created_at DESC").Find(&versions)
	if result.Error != nil {
		return nil, result.Error
	}
	return versions, nil
}

// DeleteVersion deletes a version (soft delete)
func (r *Repository) DeleteVersion(registryName, packageName, versionStr string) error {
	version, err := r.GetVersion(registryName, packageName, versionStr)
	if err != nil {
		return err
	}

	result := r.db.Delete(&models.Version{}, version.ID)
	return result.Error
}

// ==================== Advanced Queries ====================

// GetAllPackagesWithVersions returns all packages with their versions for a registry (for CDT index)
func (r *Repository) GetAllPackagesWithVersions(registryName string) ([]models.Package, error) {
	registry, err := r.GetRegistry(registryName)
	if err != nil {
		return nil, err
	}

	var packages []models.Package
	result := r.db.Preload("Versions").Where("registry_id = ?", registry.ID).Find(&packages)
	if result.Error != nil {
		return nil, result.Error
	}
	return packages, nil
}
