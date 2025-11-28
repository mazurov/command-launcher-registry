package provider

import (
	"context"

	"github.com/mazurov/command-launcher-registry/internal/auth"
)

// AuthProvider defines authentication provider interface
type AuthProvider interface {
	// Name returns provider identifier ("github", "ldap", etc.)
	Name() string

	// GetAuthURL returns OAuth authorization URL (for OAuth-based providers)
	// state parameter for CSRF protection
	GetAuthURL(state string) string

	// HandleCallback processes OAuth callback and returns authenticated user
	// code: authorization code from OAuth callback
	HandleCallback(ctx context.Context, code string) (*auth.User, error)

	// ValidateToken validates JWT token and returns user info
	// For providers that need to refresh data, this can query provider API
	ValidateToken(ctx context.Context, claims *auth.JWTClaims) (*auth.User, error)

	// Authenticate handles direct authentication (for non-OAuth providers like LDAP)
	Authenticate(ctx context.Context, username, password string) (*auth.User, error)
}
