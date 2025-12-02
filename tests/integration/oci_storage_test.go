package integration

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/criteo/command-launcher-registry/internal/models"
	"github.com/criteo/command-launcher-registry/internal/storage"
)

// newTestLogger creates a logger for integration tests
func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

// getOCITestConfig returns OCI test configuration from environment.
// Returns empty strings if not configured.
func getOCITestConfig() (uri, token string) {
	return os.Getenv("COLA_TEST_OCI_URI"), os.Getenv("COLA_TEST_OCI_TOKEN")
}

// skipIfNoOCIConfig skips the test if OCI configuration is not available.
func skipIfNoOCIConfig(t *testing.T) (uri, token string) {
	uri, token = getOCITestConfig()
	if uri == "" || token == "" {
		t.Skip("Skipping OCI integration test (set COLA_TEST_OCI_URI and COLA_TEST_OCI_TOKEN to run)")
	}
	return uri, token
}

// TestOCIStorage_Integration tests the full OCI storage lifecycle.
// Requires COLA_TEST_OCI_URI and COLA_TEST_OCI_TOKEN environment variables.
func TestOCIStorage_Integration(t *testing.T) {
	uri, token := skipIfNoOCIConfig(t)

	logger := newTestLogger()

	// Parse URI
	parsedURI, err := storage.ParseStorageURI(uri)
	require.NoError(t, err)
	require.True(t, parsedURI.IsOCIScheme())

	// Create OCI storage
	store, err := storage.NewStorage(parsedURI, token, logger)
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()

	t.Run("CreateRegistry", func(t *testing.T) {
		reg := models.NewRegistry("oci-test-reg", "OCI Test Registry", nil, nil)
		err := store.CreateRegistry(ctx, reg)
		require.NoError(t, err)
	})

	t.Run("GetRegistry", func(t *testing.T) {
		reg, err := store.GetRegistry(ctx, "oci-test-reg")
		require.NoError(t, err)
		assert.Equal(t, "oci-test-reg", reg.Name)
		assert.Equal(t, "OCI Test Registry", reg.Description)
	})

	t.Run("CreatePackage", func(t *testing.T) {
		pkg := models.NewPackage("oci-test-pkg", "OCI Test Package", nil, nil)
		err := store.CreatePackage(ctx, "oci-test-reg", pkg)
		require.NoError(t, err)
	})

	t.Run("CreateVersion", func(t *testing.T) {
		ver := &models.Version{
			Name:           "oci-test-pkg",
			Version:        "1.0.0",
			Checksum:       "abc123",
			URL:            "https://example.com/pkg.zip",
			StartPartition: 0,
			EndPartition:   9,
		}
		err := store.CreateVersion(ctx, "oci-test-reg", "oci-test-pkg", ver)
		require.NoError(t, err)
	})

	t.Run("GetRegistryIndex", func(t *testing.T) {
		entries, err := store.GetRegistryIndex(ctx, "oci-test-reg")
		require.NoError(t, err)
		assert.Len(t, entries, 1)
		assert.Equal(t, "oci-test-pkg", entries[0].Name)
		assert.Equal(t, "1.0.0", entries[0].Version)
	})

	t.Run("DeleteVersion", func(t *testing.T) {
		err := store.DeleteVersion(ctx, "oci-test-reg", "oci-test-pkg", "1.0.0")
		require.NoError(t, err)
	})

	t.Run("DeletePackage", func(t *testing.T) {
		err := store.DeletePackage(ctx, "oci-test-reg", "oci-test-pkg")
		require.NoError(t, err)
	})

	t.Run("DeleteRegistry", func(t *testing.T) {
		err := store.DeleteRegistry(ctx, "oci-test-reg")
		require.NoError(t, err)
	})
}

// TestOCIStorage_ServerStartup tests that server can start with OCI storage URI.
// This test verifies the factory correctly validates OCI configuration.
func TestOCIStorage_FactoryValidation(t *testing.T) {
	logger := newTestLogger()

	t.Run("OCIRequiresToken", func(t *testing.T) {
		uri, err := storage.ParseStorageURI("oci://ghcr.io/test/repo")
		require.NoError(t, err)

		// Creating OCI storage without token should fail
		_, err = storage.NewStorage(uri, "", logger)
		assert.Error(t, err)
		assert.ErrorIs(t, err, storage.ErrTokenRequired)
	})

	t.Run("FileDoesNotRequireToken", func(t *testing.T) {
		uri, err := storage.ParseStorageURI("file://./test-data/oci-factory-test.json")
		require.NoError(t, err)

		// Creating file storage without token should succeed
		store, err := storage.NewStorage(uri, "", logger)
		require.NoError(t, err)
		store.Close()

		// Cleanup
		os.Remove("./test-data/oci-factory-test.json")
		os.Remove("./test-data")
	})
}

// TestOCIStorage_MultiRegistry tests OCI storage with different registry providers.
// Each sub-test requires its own environment variables.
func TestOCIStorage_MultiRegistry(t *testing.T) {
	// GitHub Container Registry
	t.Run("ghcr.io", func(t *testing.T) {
		uri := os.Getenv("COLA_TEST_GHCR_URI")
		token := os.Getenv("COLA_TEST_GHCR_TOKEN")
		if uri == "" || token == "" {
			t.Skip("Skipping ghcr.io test (set COLA_TEST_GHCR_URI and COLA_TEST_GHCR_TOKEN)")
		}
		testOCIStorageBasicOps(t, uri, token)
	})

	// Docker Hub
	t.Run("docker.io", func(t *testing.T) {
		uri := os.Getenv("COLA_TEST_DOCKERHUB_URI")
		token := os.Getenv("COLA_TEST_DOCKERHUB_TOKEN")
		if uri == "" || token == "" {
			t.Skip("Skipping docker.io test (set COLA_TEST_DOCKERHUB_URI and COLA_TEST_DOCKERHUB_TOKEN)")
		}
		testOCIStorageBasicOps(t, uri, token)
	})

	// Azure Container Registry
	t.Run("azurecr.io", func(t *testing.T) {
		uri := os.Getenv("COLA_TEST_ACR_URI")
		token := os.Getenv("COLA_TEST_ACR_TOKEN")
		if uri == "" || token == "" {
			t.Skip("Skipping Azure ACR test (set COLA_TEST_ACR_URI and COLA_TEST_ACR_TOKEN)")
		}
		testOCIStorageBasicOps(t, uri, token)
	})
}

// testOCIStorageBasicOps performs basic CRUD operations on an OCI storage.
func testOCIStorageBasicOps(t *testing.T, uri, token string) {
	logger := newTestLogger()

	parsedURI, err := storage.ParseStorageURI(uri)
	require.NoError(t, err)

	store, err := storage.NewStorage(parsedURI, token, logger)
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()

	// Create and verify registry
	reg := models.NewRegistry("multi-reg-test", "Multi-Registry Test", nil, nil)
	err = store.CreateRegistry(ctx, reg)
	require.NoError(t, err)

	// Verify it was created
	retrieved, err := store.GetRegistry(ctx, "multi-reg-test")
	require.NoError(t, err)
	assert.Equal(t, "multi-reg-test", retrieved.Name)

	// Cleanup
	err = store.DeleteRegistry(ctx, "multi-reg-test")
	require.NoError(t, err)
}
