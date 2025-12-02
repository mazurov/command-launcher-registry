package storage

import (
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestOCIStorageLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestNewOCIStorage_InvalidScheme(t *testing.T) {
	logger := newTestOCIStorageLogger()

	// Parse a file URI
	uri, err := ParseStorageURI("file://./data/registry.json")
	require.NoError(t, err)

	// Try to create OCI storage with file URI - should fail
	_, err = NewOCIStorage(uri, "token", logger)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected OCI URI")
}

func TestNewOCIStorage_ValidURIParsing(t *testing.T) {
	// Skip if not running integration tests (needs real OCI registry)
	if os.Getenv("COLA_TEST_OCI_INTEGRATION") == "" {
		t.Skip("Skipping OCI storage test (set COLA_TEST_OCI_INTEGRATION=1 to run)")
	}

	logger := newTestOCIStorageLogger()

	ociURI := os.Getenv("COLA_TEST_OCI_URI")
	ociToken := os.Getenv("COLA_TEST_OCI_TOKEN")
	if ociURI == "" || ociToken == "" {
		t.Skip("Skipping OCI storage test (set COLA_TEST_OCI_URI and COLA_TEST_OCI_TOKEN)")
	}

	uri, err := ParseStorageURI(ociURI)
	require.NoError(t, err)
	require.True(t, uri.IsOCIScheme())

	storage, err := NewOCIStorage(uri, ociToken, logger)
	require.NoError(t, err)
	assert.NotNil(t, storage)
}

func TestFactory_NewStorage_FileScheme(t *testing.T) {
	logger := newTestOCIStorageLogger()

	uri, err := ParseStorageURI("file://./test-data/factory-test.json")
	require.NoError(t, err)

	store, err := NewStorage(uri, "", logger)
	require.NoError(t, err)
	assert.NotNil(t, store)

	// Clean up
	store.Close()
	os.Remove("./test-data/factory-test.json")
	os.Remove("./test-data")
}

func TestFactory_NewStorage_OCIScheme_NoToken(t *testing.T) {
	logger := newTestOCIStorageLogger()

	uri, err := ParseStorageURI("oci://ghcr.io/test/repo")
	require.NoError(t, err)

	// OCI without token should fail with ErrTokenRequired
	_, err = NewStorage(uri, "", logger)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrTokenRequired)
	assert.Contains(t, err.Error(), "OCI storage requires authentication token")
}

func TestFactory_NewStorage_UnsupportedScheme(t *testing.T) {
	logger := newTestOCIStorageLogger()

	// Manually create a URI with unsupported scheme
	uri := &StorageURI{
		Scheme: "s3",
		Path:   "bucket/path",
		Raw:    "s3://bucket/path",
	}

	_, err := NewStorage(uri, "", logger)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported storage scheme")
}

func TestFactory_NewStorage_OCIScheme_WithToken(t *testing.T) {
	// Skip if not running integration tests
	if os.Getenv("COLA_TEST_OCI_INTEGRATION") == "" {
		t.Skip("Skipping OCI factory test (set COLA_TEST_OCI_INTEGRATION=1 to run)")
	}

	logger := newTestOCIStorageLogger()

	ociURI := os.Getenv("COLA_TEST_OCI_URI")
	ociToken := os.Getenv("COLA_TEST_OCI_TOKEN")
	if ociURI == "" || ociToken == "" {
		t.Skip("Skipping (set COLA_TEST_OCI_URI and COLA_TEST_OCI_TOKEN)")
	}

	uri, err := ParseStorageURI(ociURI)
	require.NoError(t, err)

	store, err := NewStorage(uri, ociToken, logger)
	require.NoError(t, err)
	assert.NotNil(t, store)
}
