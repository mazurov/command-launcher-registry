package storage

import (
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestS3Logger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestNewS3Storage_InvalidScheme(t *testing.T) {
	logger := newTestS3Logger()

	// Create a URI with file scheme
	uri := &StorageURI{
		Scheme: "file",
		Path:   "./test/data.json",
		Raw:    "file://./test/data.json",
	}

	_, err := NewS3Storage(uri, "access:secret", logger)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expected S3 URI")
}

func TestFactory_NewStorage_S3Scheme(t *testing.T) {
	logger := newTestS3Logger()

	// Parse S3 URI
	uri, err := ParseStorageURI("s3://s3.amazonaws.com/test-bucket/registry.json")
	require.NoError(t, err)

	// Factory should route to S3Storage (will fail to connect, but tests routing)
	_, err = NewStorage(uri, "access:secret", logger)
	// Error expected because we can't connect to S3
	require.Error(t, err)
	// But it should be an S3 error, not "unsupported scheme"
	assert.NotContains(t, err.Error(), "unsupported storage scheme")
}

func TestFactory_NewStorage_S3HttpScheme(t *testing.T) {
	logger := newTestS3Logger()

	// Parse S3+HTTP URI (MinIO style)
	uri, err := ParseStorageURI("s3+http://localhost:9000/test-bucket/registry.json")
	require.NoError(t, err)

	// Factory should route to S3Storage (will fail to connect, but tests routing)
	_, err = NewStorage(uri, "access:secret", logger)
	// Error expected because we can't connect to MinIO
	require.Error(t, err)
	// But it should be an S3 error, not "unsupported scheme"
	assert.NotContains(t, err.Error(), "unsupported storage scheme")
}

func TestS3Client_TimeoutConstants(t *testing.T) {
	// Verify timeout constants
	assert.Equal(t, int64(60), int64(S3UploadTimeout.Seconds()), "Upload timeout should be 60 seconds")
	assert.Equal(t, int64(30), int64(S3DownloadTimeout.Seconds()), "Download timeout should be 30 seconds")
}
