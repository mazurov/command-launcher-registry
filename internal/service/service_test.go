package service

import (
	"strings"
	"testing"

	"github.com/mazurov/command-launcher-registry/internal/config"
	"github.com/mazurov/command-launcher-registry/internal/repository"
	"github.com/mazurov/command-launcher-registry/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestService(t *testing.T) *Service {
	// Create in-memory SQLite database for testing
	db, err := config.InitDatabase(config.DatabaseConfig{
		Type: "sqlite",
		DSN:  ":memory:",
	}, "error")
	require.NoError(t, err)

	repo := repository.NewRepository(db)
	return NewService(repo)
}

func TestCreateRegistry(t *testing.T) {
	svc := setupTestService(t)

	req := &types.CreateRegistryRequest{
		Name:        "test-registry",
		Description: "Test Registry",
		Admin:       []string{"admin@test.com"},
	}

	registry, err := svc.CreateRegistry(req)
	require.NoError(t, err)
	assert.Equal(t, "test-registry", registry.Name)
	assert.Equal(t, "Test Registry", registry.Description)
}

func TestCreatePackage(t *testing.T) {
	svc := setupTestService(t)

	// First create a registry
	regReq := &types.CreateRegistryRequest{
		Name:        "test-registry",
		Description: "Test Registry",
	}
	_, err := svc.CreateRegistry(regReq)
	require.NoError(t, err)

	// Create a package
	pkgReq := &types.CreatePackageRequest{
		Name:        "test-package",
		Description: "Test Package",
	}

	pkg, err := svc.CreatePackage("test-registry", pkgReq)
	require.NoError(t, err)
	assert.Equal(t, "test-package", pkg.Name)
	assert.Equal(t, "Test Package", pkg.Description)
}

func TestPublishVersion(t *testing.T) {
	svc := setupTestService(t)

	// Setup registry and package
	regReq := &types.CreateRegistryRequest{Name: "test-registry"}
	_, err := svc.CreateRegistry(regReq)
	require.NoError(t, err)

	pkgReq := &types.CreatePackageRequest{Name: "test-package"}
	_, err = svc.CreatePackage("test-registry", pkgReq)
	require.NoError(t, err)

	// Publish version
	verReq := &types.PublishVersionRequest{
		Version:  "1.0.0",
		URL:      "http://example.com/pkg.zip",
		Checksum: "abc123",
	}

	version, err := svc.PublishVersion("test-registry", "test-package", verReq)
	require.NoError(t, err)
	assert.Equal(t, "1.0.0", version.Version)
	assert.Equal(t, "http://example.com/pkg.zip", version.URL)
	assert.Equal(t, uint8(0), version.StartPartition)
	assert.Equal(t, uint8(9), version.EndPartition)
}

func TestGetRegistryIndex(t *testing.T) {
	svc := setupTestService(t)

	// Setup test data
	regReq := &types.CreateRegistryRequest{Name: "test-registry"}
	_, err := svc.CreateRegistry(regReq)
	require.NoError(t, err)

	pkgReq := &types.CreatePackageRequest{Name: "test-package"}
	_, err = svc.CreatePackage("test-registry", pkgReq)
	require.NoError(t, err)

	verReq := &types.PublishVersionRequest{
		Version:  "1.0.0",
		URL:      "http://example.com/pkg.zip",
		Checksum: "abc123",
	}
	_, err = svc.PublishVersion("test-registry", "test-package", verReq)
	require.NoError(t, err)

	// Get index (flat list)
	index, err := svc.GetRegistryIndex("test-registry")
	require.NoError(t, err)
	assert.NotNil(t, index)
	assert.Len(t, index, 1)
	assert.Equal(t, "test-package", index[0].Name)
	assert.Equal(t, "1.0.0", index[0].Version)
}

func TestDuplicateRegistry(t *testing.T) {
	svc := setupTestService(t)

	req := &types.CreateRegistryRequest{Name: "test-registry"}
	_, err := svc.CreateRegistry(req)
	require.NoError(t, err)

	// Try to create again - should fail with duplicate error
	_, err = svc.CreateRegistry(req)
	assert.Error(t, err)
	// SQLite returns "UNIQUE constraint failed" rather than a custom message
	errMsg := err.Error()
	isDuplicate := strings.Contains(errMsg, "UNIQUE") || strings.Contains(errMsg, "already exists")
	assert.True(t, isDuplicate, "Expected duplicate error, got: %s", errMsg)
}
