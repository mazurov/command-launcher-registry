package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseS3Token_ValidToken(t *testing.T) {
	accessKey, secretKey, err := ParseS3Token("AKIAIOSFODNN7EXAMPLE:wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY")
	require.NoError(t, err)
	assert.Equal(t, "AKIAIOSFODNN7EXAMPLE", accessKey)
	assert.Equal(t, "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", secretKey)
}

func TestParseS3Token_SecretWithColons(t *testing.T) {
	// Secret key may contain colons - only split on first colon
	accessKey, secretKey, err := ParseS3Token("ACCESS:SECRET:WITH:MULTIPLE:COLONS")
	require.NoError(t, err)
	assert.Equal(t, "ACCESS", accessKey)
	assert.Equal(t, "SECRET:WITH:MULTIPLE:COLONS", secretKey)
}

func TestParseS3Token_InvalidFormat(t *testing.T) {
	_, _, err := ParseS3Token("invalid-no-colon")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid token format")
}

func TestParseS3Token_EmptyAccessKey(t *testing.T) {
	_, _, err := ParseS3Token(":secretkey")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "access key cannot be empty")
}

func TestParseS3Token_EmptySecretKey(t *testing.T) {
	_, _, err := ParseS3Token("accesskey:")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "secret key cannot be empty")
}

func TestExtractRegionFromEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		expected string
	}{
		{
			name:     "AWS S3 dot format",
			endpoint: "s3.us-west-2.amazonaws.com",
			expected: "us-west-2",
		},
		{
			name:     "AWS S3 hyphen format",
			endpoint: "s3-us-east-1.amazonaws.com",
			expected: "us-east-1",
		},
		{
			name:     "AWS S3 eu-central-1",
			endpoint: "s3.eu-central-1.amazonaws.com",
			expected: "eu-central-1",
		},
		{
			name:     "MinIO - no region",
			endpoint: "minio.example.com:9000",
			expected: "",
		},
		{
			name:     "DigitalOcean Spaces - no region extraction",
			endpoint: "nyc3.digitaloceanspaces.com",
			expected: "",
		},
		{
			name:     "Global S3 endpoint",
			endpoint: "s3.amazonaws.com",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractRegionFromEndpoint(tt.endpoint)
			assert.Equal(t, tt.expected, result)
		})
	}
}
