package service

import (
	"fmt"

	"github.com/mazurov/command-launcher-registry/internal/models"
	"github.com/mazurov/command-launcher-registry/internal/repository"
	"github.com/mazurov/command-launcher-registry/pkg/types"
)

// Service provides business logic
type Service struct {
	repo *repository.Repository
}

// NewService creates a new service instance
func NewService(repo *repository.Repository) *Service {
	return &Service{repo: repo}
}

// ==================== Registry Operations ====================

// CreateRegistry creates a new registry
func (s *Service) CreateRegistry(req *types.CreateRegistryRequest) (*models.Registry, error) {
	registry := &models.Registry{
		Name:         req.Name,
		Description:  req.Description,
		Admin:        req.Admin,
		CustomValues: req.CustomValues,
	}

	if err := s.repo.CreateRegistry(registry); err != nil {
		return nil, err
	}

	return registry, nil
}

// GetRegistry retrieves a registry by name
func (s *Service) GetRegistry(name string) (*models.Registry, error) {
	return s.repo.GetRegistry(name)
}

// ListRegistries returns all registries
func (s *Service) ListRegistries() ([]models.Registry, error) {
	return s.repo.ListRegistries()
}

// UpdateRegistry updates an existing registry
func (s *Service) UpdateRegistry(name string, req *types.UpdateRegistryRequest) error {
	updates := make(map[string]interface{})
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Admin != nil {
		updates["admin"] = req.Admin
	}
	if req.CustomValues != nil {
		updates["custom_values"] = req.CustomValues
	}

	return s.repo.UpdateRegistry(name, updates)
}

// DeleteRegistry deletes a registry
func (s *Service) DeleteRegistry(name string) error {
	return s.repo.DeleteRegistry(name)
}

// ==================== Package Operations ====================

// CreatePackage creates a new package within a registry
func (s *Service) CreatePackage(registryName string, req *types.CreatePackageRequest) (*models.Package, error) {
	pkg := &models.Package{
		Name:         req.Name,
		Description:  req.Description,
		Admin:        req.Admin,
		CustomValues: req.CustomValues,
	}

	if err := s.repo.CreatePackage(registryName, pkg); err != nil {
		return nil, err
	}

	return pkg, nil
}

// GetPackage retrieves a package
func (s *Service) GetPackage(registryName, packageName string) (*models.Package, error) {
	return s.repo.GetPackage(registryName, packageName)
}

// ListPackages returns all packages in a registry
func (s *Service) ListPackages(registryName string) ([]models.Package, error) {
	return s.repo.ListPackages(registryName)
}

// UpdatePackage updates an existing package
func (s *Service) UpdatePackage(registryName, packageName string, req *types.UpdatePackageRequest) error {
	updates := make(map[string]interface{})
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Admin != nil {
		updates["admin"] = req.Admin
	}
	if req.CustomValues != nil {
		updates["custom_values"] = req.CustomValues
	}

	return s.repo.UpdatePackage(registryName, packageName, updates)
}

// DeletePackage deletes a package
func (s *Service) DeletePackage(registryName, packageName string) error {
	return s.repo.DeletePackage(registryName, packageName)
}

// ==================== Version Operations ====================

// PublishVersion publishes a new version of a package
func (s *Service) PublishVersion(registryName, packageName string, req *types.PublishVersionRequest) (*models.Version, error) {
	// Set defaults for partitions if not provided
	startPartition := req.StartPartition
	endPartition := req.EndPartition
	if endPartition == 0 && startPartition == 0 {
		endPartition = 9
	}

	version := &models.Version{
		Version:        req.Version,
		URL:            req.URL,
		Checksum:       req.Checksum,
		StartPartition: startPartition,
		EndPartition:   endPartition,
	}

	if err := s.repo.CreateVersion(registryName, packageName, version); err != nil {
		return nil, err
	}

	return version, nil
}

// GetVersion retrieves a specific version
func (s *Service) GetVersion(registryName, packageName, version string) (*models.Version, error) {
	return s.repo.GetVersion(registryName, packageName, version)
}

// ListVersions returns all versions of a package
func (s *Service) ListVersions(registryName, packageName string) ([]models.Version, error) {
	return s.repo.ListVersions(registryName, packageName)
}

// DeleteVersion deletes a version
func (s *Service) DeleteVersion(registryName, packageName, version string) error {
	return s.repo.DeleteVersion(registryName, packageName, version)
}

// ==================== CDT-Compatible Index Generation ====================

// GetRegistryIndex generates a CDT-compatible registry index (flat list of versions)
func (s *Service) GetRegistryIndex(registryName string) ([]types.PackageInfo, error) {
	packages, err := s.repo.GetAllPackagesWithVersions(registryName)
	if err != nil {
		return nil, err
	}

	// Flatten all versions from all packages into a single list
	// Initialize with empty slice to ensure JSON returns [] instead of null
	versions := make([]types.PackageInfo, 0)
	for _, pkg := range packages {
		for _, v := range pkg.Versions {
			versions = append(versions, types.PackageInfo{
				Name:           pkg.Name,
				Version:        v.Version,
				URL:            v.URL,
				Checksum:       v.Checksum,
				StartPartition: v.StartPartition,
				EndPartition:   v.EndPartition,
			})
		}
	}

	return versions, nil
}

// ==================== Validation & Helpers ====================

// ValidateVersionFormat validates version string format
func (s *Service) ValidateVersionFormat(version string) error {
	// Basic validation - can be enhanced with semver library
	if version == "" {
		return fmt.Errorf("version cannot be empty")
	}
	// Add more sophisticated validation if needed
	return nil
}

// ValidateURL validates URL format
func (s *Service) ValidateURL(url string) error {
	if url == "" {
		return fmt.Errorf("URL cannot be empty")
	}
	// Add URL format validation if needed
	return nil
}
