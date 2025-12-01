package auth

import (
	"testing"
)

// TODO: Add comprehensive unit tests for credential storage
// Test cases:
// - SaveCredentials and LoadCredentials roundtrip
// - DeleteCredentials removes all stored data
// - LoadStoredToken returns ErrNotFound when no credentials exist
// - LoadStoredURL returns ErrNotFound when no credentials exist
// - File permissions are set correctly (0600)

func TestCredentialStorage(t *testing.T) {
	t.Skip("TODO: Implement credential storage tests")
}
