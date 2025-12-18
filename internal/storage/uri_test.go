package storage

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeStorageURI(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "path without scheme",
			input:    "./data/registry.json",
			expected: "file://./data/registry.json",
		},
		{
			name:     "absolute path without scheme",
			input:    "/var/data/registry.json",
			expected: "file:///var/data/registry.json",
		},
		{
			name:     "already has file scheme",
			input:    "file://./data/registry.json",
			expected: "file://./data/registry.json",
		},
		{
			name:     "oci scheme unchanged",
			input:    "oci://registry.example.com/repo",
			expected: "oci://registry.example.com/repo",
		},
		{
			name:     "windows path without scheme",
			input:    "C:/data/registry.json",
			expected: "file://C:/data/registry.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeStorageURI(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseStorageURI_ValidURIs(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedScheme string
		expectedPath   string
	}{
		{
			name:           "file URI with relative path",
			input:          "file://./data/registry.json",
			expectedScheme: "file",
			expectedPath:   "./data/registry.json",
		},
		{
			name:           "file URI with absolute path",
			input:          "file:///var/data/registry.json",
			expectedScheme: "file",
			expectedPath:   "/var/data/registry.json",
		},
		{
			name:           "path without scheme (auto-prefixed)",
			input:          "./data/registry.json",
			expectedScheme: "file",
			expectedPath:   "./data/registry.json",
		},
		{
			name:           "absolute path without scheme",
			input:          "/var/data/registry.json",
			expectedScheme: "file",
			expectedPath:   "/var/data/registry.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uri, err := ParseStorageURI(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedScheme, uri.Scheme)
			assert.Equal(t, tt.expectedPath, uri.Path)
			assert.Equal(t, tt.input, uri.Raw)
		})
	}
}

func TestParseStorageURI_InvalidURIs(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		errContains string
	}{
		{
			name:        "empty URI",
			input:       "",
			errContains: "cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseStorageURI(tt.input)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errContains)
		})
	}
}

func TestParseStorageURI_ValidOCIURIs(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedScheme string
		expectedHost   string
		expectedPath   string
	}{
		{
			name:           "ghcr.io repository",
			input:          "oci://ghcr.io/myorg/cola-data",
			expectedScheme: "oci",
			expectedHost:   "ghcr.io",
			expectedPath:   "myorg/cola-data",
		},
		{
			name:           "docker.io repository",
			input:          "oci://docker.io/user/cola-data",
			expectedScheme: "oci",
			expectedHost:   "docker.io",
			expectedPath:   "user/cola-data",
		},
		{
			name:           "azure container registry",
			input:          "oci://myregistry.azurecr.io/cola/data",
			expectedScheme: "oci",
			expectedHost:   "myregistry.azurecr.io",
			expectedPath:   "cola/data",
		},
		{
			name:           "with tag (tag stripped)",
			input:          "oci://ghcr.io/myorg/cola-data:v1.0",
			expectedScheme: "oci",
			expectedHost:   "ghcr.io",
			expectedPath:   "myorg/cola-data",
		},
		{
			name:           "deep path",
			input:          "oci://registry.example.com/org/team/project/data",
			expectedScheme: "oci",
			expectedHost:   "registry.example.com",
			expectedPath:   "org/team/project/data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uri, err := ParseStorageURI(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedScheme, uri.Scheme)
			assert.Equal(t, tt.expectedHost, uri.Host)
			assert.Equal(t, tt.expectedPath, uri.Path)
			assert.Equal(t, tt.input, uri.Raw)
		})
	}
}

func TestParseStorageURI_InvalidOCIURIs(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		errContains string
	}{
		{
			name:        "no host",
			input:       "oci:///path",
			errContains: "OCI URI must include registry host",
		},
		{
			name:        "no path",
			input:       "oci://ghcr.io",
			errContains: "OCI URI must include repository path",
		},
		{
			name:        "with query params",
			input:       "oci://ghcr.io/repo?foo=bar",
			errContains: "OCI URI does not support query parameters",
		},
		{
			name:        "with fragment",
			input:       "oci://ghcr.io/repo#section",
			errContains: "OCI URI does not support fragments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseStorageURI(tt.input)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errContains)
		})
	}
}

func TestParseStorageURI_UnknownScheme(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		scheme string
	}{
		{
			name:   "http scheme",
			input:  "http://example.com/path",
			scheme: "http",
		},
		{
			name:   "custom scheme",
			input:  "custom://host/path",
			scheme: "custom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseStorageURI(tt.input)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "unsupported storage scheme")
			assert.Contains(t, err.Error(), tt.scheme)
			assert.Contains(t, err.Error(), "supported schemes")
			// Verify supported schemes are listed
			assert.Contains(t, err.Error(), "file")
		})
	}
}

func TestParseStorageURI_SupportedSchemesListed(t *testing.T) {
	_, err := ParseStorageURI("unknown://path")
	require.Error(t, err)

	// Verify all supported schemes are mentioned in error
	for _, scheme := range SupportedSchemes {
		assert.True(t, strings.Contains(err.Error(), scheme),
			"Error should list supported scheme: %s", scheme)
	}
}

func TestStorageURI_IsFileScheme(t *testing.T) {
	fileURI, err := ParseStorageURI("file://./data/registry.json")
	require.NoError(t, err)
	assert.True(t, fileURI.IsFileScheme())
	assert.False(t, fileURI.IsOCIScheme())
}

func TestStorageURI_IsOCIScheme(t *testing.T) {
	ociURI, err := ParseStorageURI("oci://ghcr.io/myorg/cola-data")
	require.NoError(t, err)
	assert.True(t, ociURI.IsOCIScheme())
	assert.False(t, ociURI.IsFileScheme())
}

func TestStorageURI_OCIReference(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedRef string
	}{
		{
			name:        "simple path",
			input:       "oci://ghcr.io/myorg/cola-data",
			expectedRef: "ghcr.io/myorg/cola-data:latest",
		},
		{
			name:        "deep path",
			input:       "oci://registry.example.com/org/team/project",
			expectedRef: "registry.example.com/org/team/project:latest",
		},
		{
			name:        "with tag stripped",
			input:       "oci://ghcr.io/myorg/cola-data:v1.0",
			expectedRef: "ghcr.io/myorg/cola-data:latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uri, err := ParseStorageURI(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedRef, uri.OCIReference())
		})
	}
}

func TestStorageURI_String(t *testing.T) {
	input := "./data/registry.json"
	uri, err := ParseStorageURI(input)
	require.NoError(t, err)
	assert.Equal(t, input, uri.String())
}

// S3 URI Tests (T1.3)

func TestParseStorageURI_ValidS3URIs(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedScheme string
		expectedHost   string
		expectedBucket string
		expectedKey    string
		expectedRegion string
		expectedSSL    bool
	}{
		{
			name:           "AWS S3 with endpoint",
			input:          "s3://s3.amazonaws.com/my-bucket/path/to/registry.json",
			expectedScheme: "s3",
			expectedHost:   "s3.amazonaws.com",
			expectedBucket: "my-bucket",
			expectedKey:    "path/to/registry.json",
			expectedRegion: "",
			expectedSSL:    true,
		},
		{
			name:           "AWS S3 with region in query",
			input:          "s3://s3.amazonaws.com/my-bucket/registry.json?region=us-east-1",
			expectedScheme: "s3",
			expectedHost:   "s3.amazonaws.com",
			expectedBucket: "my-bucket",
			expectedKey:    "registry.json",
			expectedRegion: "us-east-1",
			expectedSSL:    true,
		},
		{
			name:           "MinIO HTTP scheme",
			input:          "s3+http://localhost:9000/cola-bucket/registry.json",
			expectedScheme: "s3+http",
			expectedHost:   "localhost:9000",
			expectedBucket: "cola-bucket",
			expectedKey:    "registry.json",
			expectedRegion: "",
			expectedSSL:    false,
		},
		{
			name:           "DigitalOcean Spaces",
			input:          "s3://nyc3.digitaloceanspaces.com/my-space/cola/registry.json",
			expectedScheme: "s3",
			expectedHost:   "nyc3.digitaloceanspaces.com",
			expectedBucket: "my-space",
			expectedKey:    "cola/registry.json",
			expectedRegion: "",
			expectedSSL:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uri, err := ParseStorageURI(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedScheme, uri.Scheme)
			assert.Equal(t, tt.expectedHost, uri.Host)
			assert.True(t, uri.IsS3Scheme())
			assert.Equal(t, tt.expectedBucket, uri.S3Bucket())
			assert.Equal(t, tt.expectedKey, uri.S3Key())
			assert.Equal(t, tt.expectedRegion, uri.S3Region())
			assert.Equal(t, tt.expectedSSL, uri.S3UseSSL())
		})
	}
}

func TestParseStorageURI_InvalidS3URIs(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		errContains string
	}{
		{
			name:        "no host",
			input:       "s3:///bucket/path",
			errContains: "S3 URI must include endpoint host",
		},
		{
			name:        "no path",
			input:       "s3://s3.amazonaws.com",
			errContains: "S3 URI must include bucket and path",
		},
		{
			name:        "only bucket, no key",
			input:       "s3://s3.amazonaws.com/bucket",
			errContains: "S3 URI must include object key path",
		},
		{
			name:        "with fragment",
			input:       "s3://s3.amazonaws.com/bucket/path#section",
			errContains: "S3 URI does not support fragments",
		},
		{
			name:        "unknown query param",
			input:       "s3://s3.amazonaws.com/bucket/path?foo=bar",
			errContains: "S3 URI does not support query parameter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseStorageURI(tt.input)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errContains)
		})
	}
}

func TestStorageURI_IsS3Scheme(t *testing.T) {
	s3URI, err := ParseStorageURI("s3://s3.amazonaws.com/bucket/registry.json")
	require.NoError(t, err)
	assert.True(t, s3URI.IsS3Scheme())
	assert.False(t, s3URI.IsFileScheme())
	assert.False(t, s3URI.IsOCIScheme())

	s3HttpURI, err := ParseStorageURI("s3+http://localhost:9000/bucket/registry.json")
	require.NoError(t, err)
	assert.True(t, s3HttpURI.IsS3Scheme())
	assert.False(t, s3HttpURI.S3UseSSL())
}
