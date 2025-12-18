package integration

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/criteo/command-launcher-registry/internal/models"
	"github.com/criteo/command-launcher-registry/internal/storage"
)

// S3 integration test environment variables
const (
	envS3Endpoint  = "COLA_TEST_S3_ENDPOINT"
	envS3Bucket    = "COLA_TEST_S3_BUCKET"
	envS3AccessKey = "COLA_TEST_S3_ACCESS_KEY"
	envS3SecretKey = "COLA_TEST_S3_SECRET_KEY"
	envS3UseSSL    = "COLA_TEST_S3_USE_SSL" // "true" or "false", default "true"
)

func skipIfNoS3(t *testing.T) {
	if os.Getenv(envS3Endpoint) == "" {
		t.Skipf("S3 integration tests require %s environment variable", envS3Endpoint)
	}
	if os.Getenv(envS3Bucket) == "" {
		t.Skipf("S3 integration tests require %s environment variable", envS3Bucket)
	}
	if os.Getenv(envS3AccessKey) == "" {
		t.Skipf("S3 integration tests require %s environment variable", envS3AccessKey)
	}
	if os.Getenv(envS3SecretKey) == "" {
		t.Skipf("S3 integration tests require %s environment variable", envS3SecretKey)
	}
}

func newS3TestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
}

func getS3TestConfig() (string, string) {
	endpoint := os.Getenv(envS3Endpoint)
	bucket := os.Getenv(envS3Bucket)
	accessKey := os.Getenv(envS3AccessKey)
	secretKey := os.Getenv(envS3SecretKey)
	useSSL := os.Getenv(envS3UseSSL) != "false"

	// Generate unique object key for test isolation
	testKey := fmt.Sprintf("test/registry-%d.json", time.Now().UnixNano())

	scheme := "s3"
	if !useSSL {
		scheme = "s3+http"
	}

	uri := fmt.Sprintf("%s://%s/%s/%s", scheme, endpoint, bucket, testKey)
	token := fmt.Sprintf("%s:%s", accessKey, secretKey)

	return uri, token
}

// TestS3Storage_FullCRUDLifecycle tests the complete lifecycle with S3 storage
func TestS3Storage_FullCRUDLifecycle(t *testing.T) {
	skipIfNoS3(t)

	logger := newS3TestLogger()
	uri, token := getS3TestConfig()

	t.Logf("Testing with S3 URI: %s", uri)

	// Parse URI
	storageURI, err := storage.ParseStorageURI(uri)
	require.NoError(t, err, "Failed to parse S3 URI")

	// Create storage
	store, err := storage.NewS3Storage(storageURI, token, logger)
	require.NoError(t, err, "Failed to create S3 storage")
	defer store.Close()

	ctx := context.Background()

	// Test: Create registry
	registry := &models.Registry{
		Name:        "test-registry",
		Description: "Test registry for S3 integration",
		Admins:      []string{"admin@test.com"},
		Packages:    make(map[string]*models.Package),
	}

	err = store.CreateRegistry(ctx, registry)
	require.NoError(t, err, "Failed to create registry")

	// Test: Get registry
	retrieved, err := store.GetRegistry(ctx, "test-registry")
	require.NoError(t, err, "Failed to get registry")
	assert.Equal(t, "test-registry", retrieved.Name)
	assert.Equal(t, "Test registry for S3 integration", retrieved.Description)

	// Test: Create package
	pkg := &models.Package{
		Name:        "test-package",
		Description: "Test package",
		Maintainers: []string{"maintainer@test.com"},
		Versions:    make(map[string]*models.Version),
	}

	err = store.CreatePackage(ctx, "test-registry", pkg)
	require.NoError(t, err, "Failed to create package")

	// Test: Get package
	retrievedPkg, err := store.GetPackage(ctx, "test-registry", "test-package")
	require.NoError(t, err, "Failed to get package")
	assert.Equal(t, "test-package", retrievedPkg.Name)

	// Test: Create version
	version := &models.Version{
		Name:           "test-package",
		Version:        "1.0.0",
		Checksum:       "sha256:abc123",
		URL:            "https://example.com/test-package-1.0.0.tar.gz",
		StartPartition: 0,
		EndPartition:   9,
	}

	err = store.CreateVersion(ctx, "test-registry", "test-package", version)
	require.NoError(t, err, "Failed to create version")

	// Test: Get version
	retrievedVer, err := store.GetVersion(ctx, "test-registry", "test-package", "1.0.0")
	require.NoError(t, err, "Failed to get version")
	assert.Equal(t, "1.0.0", retrievedVer.Version)
	assert.Equal(t, "sha256:abc123", retrievedVer.Checksum)

	// Test: List registries
	registries, err := store.ListRegistries(ctx)
	require.NoError(t, err, "Failed to list registries")
	assert.Len(t, registries, 1)

	// Test: Delete version
	err = store.DeleteVersion(ctx, "test-registry", "test-package", "1.0.0")
	require.NoError(t, err, "Failed to delete version")

	// Test: Delete package
	err = store.DeletePackage(ctx, "test-registry", "test-package")
	require.NoError(t, err, "Failed to delete package")

	// Test: Delete registry
	err = store.DeleteRegistry(ctx, "test-registry")
	require.NoError(t, err, "Failed to delete registry")

	// Verify cleanup
	registries, err = store.ListRegistries(ctx)
	require.NoError(t, err, "Failed to list registries after cleanup")
	assert.Len(t, registries, 0)

	t.Log("S3 full CRUD lifecycle test completed successfully")
}
