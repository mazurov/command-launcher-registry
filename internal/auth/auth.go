package auth

import (
	"net/http"
)

// User represents an authenticated user
type User struct {
	Username string
}

// Authenticator defines the authentication interface
type Authenticator interface {
	// Authenticate validates request credentials and returns user info
	Authenticate(r *http.Request) (*User, error)

	// Middleware returns HTTP middleware for the auth method
	Middleware() func(http.Handler) http.Handler
}
