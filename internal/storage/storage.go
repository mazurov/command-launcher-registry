package storage

import (
	"context"
	"errors"

	"github.com/criteo/command-launcher-registry/internal/models"
)

var (
	// ErrNotFound is returned when a resource is not found
	ErrNotFound = errors.New("resource not found")

	// ErrAlreadyExists is returned when attempting to create a resource that already exists
	ErrAlreadyExists = errors.New("resource already exists")

	// ErrStorageUnavailable is returned when storage operations fail
	ErrStorageUnavailable = errors.New("storage unavailable")

	// ErrImmutabilityViolation is returned when attempting to modify immutable data
	ErrImmutabilityViolation = errors.New("immutability violation")

	// ErrPartitionOverlap is returned when version partition ranges overlap
	ErrPartitionOverlap = errors.New("partition ranges overlap")
)

// Store defines the interface for storage operations
type Store interface {
	// Registry operations
	CreateRegistry(ctx context.Context, r *models.Registry) error
	GetRegistry(ctx context.Context, name string) (*models.Registry, error)
	UpdateRegistry(ctx context.Context, r *models.Registry) error
	DeleteRegistry(ctx context.Context, name string) error
	ListRegistries(ctx context.Context) ([]*models.Registry, error)

	// Package operations
	CreatePackage(ctx context.Context, registryName string, p *models.Package) error
	GetPackage(ctx context.Context, registryName, packageName string) (*models.Package, error)
	UpdatePackage(ctx context.Context, registryName string, p *models.Package) error
	DeletePackage(ctx context.Context, registryName, packageName string) error
	ListPackages(ctx context.Context, registryName string) ([]*models.Package, error)

	// Version operations
	CreateVersion(ctx context.Context, registryName, packageName string, v *models.Version) error
	GetVersion(ctx context.Context, registryName, packageName, version string) (*models.Version, error)
	DeleteVersion(ctx context.Context, registryName, packageName, version string) error
	ListVersions(ctx context.Context, registryName, packageName string) ([]*models.Version, error)

	// Index generation
	GetRegistryIndex(ctx context.Context, registryName string) ([]models.IndexEntry, error)

	// Close closes the storage
	Close() error
}
