package storage

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestOCILogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestNewOCIClient_ValidReference(t *testing.T) {
	logger := newTestOCILogger()

	client, err := NewOCIClient("ghcr.io/test/repo:latest", "test-token", logger)
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "ghcr.io/test/repo:latest", client.reference)
}

func TestNewOCIClient_NoToken(t *testing.T) {
	logger := newTestOCILogger()

	client, err := NewOCIClient("ghcr.io/test/repo:latest", "", logger)
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestNewOCIClient_InvalidReference(t *testing.T) {
	logger := newTestOCILogger()

	// Empty reference should fail
	_, err := NewOCIClient("", "token", logger)
	assert.Error(t, err)
}

func TestOCIClient_TimeoutConstants(t *testing.T) {
	// Verify timeout constants per FR-016
	assert.Equal(t, 5*time.Second, OCIPushTimeout, "Push timeout should be 5 seconds")
	assert.Equal(t, 30*time.Second, OCIPullTimeout, "Pull timeout should be 30 seconds")
}

func TestOCIClient_MediaTypeConstants(t *testing.T) {
	// Verify media types
	assert.Equal(t, "application/vnd.oci.image.config.v1+json", OCIConfigMediaType)
	assert.Equal(t, "application/json", OCILayerMediaType)
	assert.Equal(t, "registry.json", OCIManifestTitle)
}

// TestOCIClient_Exists_NotFound tests that non-existent artifacts return false, not error
// This test requires network access to a registry, skip if not in integration test mode
func TestOCIClient_Exists_NetworkError(t *testing.T) {
	// Skip if not running integration tests
	if os.Getenv("COLA_TEST_OCI_INTEGRATION") == "" {
		t.Skip("Skipping OCI integration test (set COLA_TEST_OCI_INTEGRATION=1 to run)")
	}

	logger := newTestOCILogger()

	// Use a non-existent registry to test network error handling
	client, err := NewOCIClient("nonexistent.invalid/test/repo:latest", "", logger)
	require.NoError(t, err)

	ctx := context.Background()
	exists, err := client.Exists(ctx)

	// Should return error for network issues (DNS failure)
	assert.Error(t, err)
	assert.False(t, exists)

	// Error should be categorized as network error
	var ociErr *OCIError
	assert.ErrorAs(t, err, &ociErr)
}

// TestOCIClient_RealRegistry tests with actual OCI registry
// Requires COLA_TEST_OCI_URI and COLA_TEST_OCI_TOKEN environment variables
func TestOCIClient_RealRegistry(t *testing.T) {
	ociURI := os.Getenv("COLA_TEST_OCI_URI")
	ociToken := os.Getenv("COLA_TEST_OCI_TOKEN")

	if ociURI == "" || ociToken == "" {
		t.Skip("Skipping OCI integration test (set COLA_TEST_OCI_URI and COLA_TEST_OCI_TOKEN to run)")
	}

	logger := newTestOCILogger()

	// Parse URI to get reference
	uri, err := ParseStorageURI(ociURI)
	require.NoError(t, err)
	require.True(t, uri.IsOCIScheme())

	reference := uri.OCIReference()
	client, err := NewOCIClient(reference, ociToken, logger)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("Push", func(t *testing.T) {
		testData := []byte(`{"registries":{}}`)
		err := client.Push(ctx, testData)
		require.NoError(t, err)
	})

	t.Run("Exists", func(t *testing.T) {
		exists, err := client.Exists(ctx)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("Pull", func(t *testing.T) {
		data, err := client.Pull(ctx)
		require.NoError(t, err)
		assert.Contains(t, string(data), "registries")
	})
}

// TestOCIClient_ErrorCategorization tests that errors are properly categorized
func TestOCIClient_ErrorCategorization(t *testing.T) {
	logger := newTestOCILogger()

	// Create client with valid reference but no actual registry
	client, err := NewOCIClient("localhost:5000/test/repo:latest", "test-token", logger)
	require.NoError(t, err)

	// Skip if we can't test network errors (localhost:5000 might exist in some environments)
	if os.Getenv("COLA_TEST_OCI_INTEGRATION") == "" {
		t.Skip("Skipping OCI error categorization test (set COLA_TEST_OCI_INTEGRATION=1 to run)")
	}

	ctx := context.Background()

	// Test that connection failures are properly categorized
	_, err = client.Exists(ctx)
	if err != nil {
		var ociErr *OCIError
		if assert.ErrorAs(t, err, &ociErr) {
			// Should be either network or storage error (connection refused)
			assert.Contains(t, []string{OCICategoryNetwork, OCICategoryStorage}, ociErr.Category)
		}
	}
}

// TestOCIClient_Push_ContextCancellation tests that push respects context cancellation
func TestOCIClient_Push_ContextCancellation(t *testing.T) {
	logger := newTestOCILogger()

	client, err := NewOCIClient("ghcr.io/test/repo:latest", "test-token", logger)
	require.NoError(t, err)

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = client.Push(ctx, []byte(`{}`))
	assert.Error(t, err)
}

// TestOCIClient_Pull_ContextCancellation tests that pull respects context cancellation
func TestOCIClient_Pull_ContextCancellation(t *testing.T) {
	logger := newTestOCILogger()

	client, err := NewOCIClient("ghcr.io/test/repo:latest", "test-token", logger)
	require.NoError(t, err)

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = client.Pull(ctx)
	assert.Error(t, err)
}

// TestOCIClient_Exists_ContextCancellation tests that exists respects context cancellation
func TestOCIClient_Exists_ContextCancellation(t *testing.T) {
	logger := newTestOCILogger()

	client, err := NewOCIClient("ghcr.io/test/repo:latest", "test-token", logger)
	require.NoError(t, err)

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = client.Exists(ctx)
	assert.Error(t, err)
}
