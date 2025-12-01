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

func TestParseStorageURI_OCINotImplemented(t *testing.T) {
	uri := "oci://registry.example.com/myrepo"
	_, err := ParseStorageURI(uri)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not yet implemented")
	assert.Contains(t, err.Error(), "oci")
	assert.Contains(t, err.Error(), "supported schemes")
	assert.Contains(t, err.Error(), "file")
}

func TestParseStorageURI_UnknownScheme(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		scheme string
	}{
		{
			name:   "s3 scheme",
			input:  "s3://bucket/path",
			scheme: "s3",
		},
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
}

func TestStorageURI_String(t *testing.T) {
	input := "./data/registry.json"
	uri, err := ParseStorageURI(input)
	require.NoError(t, err)
	assert.Equal(t, input, uri.String())
}
