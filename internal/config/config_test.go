package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaskToken(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected string
	}{
		{
			name:     "empty token",
			token:    "",
			expected: "",
		},
		{
			name:     "non-empty token",
			token:    "my-secret-token",
			expected: "***",
		},
		{
			name:     "short token",
			token:    "x",
			expected: "***",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Storage: StorageConfig{
					Token: tt.token,
				},
			}
			assert.Equal(t, tt.expected, cfg.MaskToken())
		})
	}
}

func TestValidate_StorageURI(t *testing.T) {
	tests := []struct {
		name      string
		uri       string
		wantError bool
		errMsg    string
	}{
		{
			name:      "valid file URI",
			uri:       "file://./data/registry.json",
			wantError: false,
		},
		{
			name:      "path without scheme (auto-prefixed)",
			uri:       "./data/registry.json",
			wantError: false,
		},
		{
			name:      "unsupported scheme",
			uri:       "s3://bucket/path",
			wantError: true,
			errMsg:    "unsupported storage scheme",
		},
		{
			name:      "oci scheme not yet implemented",
			uri:       "oci://registry.example.com/repo",
			wantError: true,
			errMsg:    "not yet implemented",
		},
		{
			name:      "empty URI",
			uri:       "",
			wantError: true,
			errMsg:    "cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Server: ServerConfig{
					Port: 8080,
					Host: "0.0.0.0",
				},
				Storage: StorageConfig{
					URI: tt.uri,
				},
				Auth: AuthConfig{
					Type: "none",
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
				},
			}
			err := cfg.Validate()
			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
