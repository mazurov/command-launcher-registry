package auth

import (
	"net/http"
)

// NoAuth implements Authenticator with no authentication (all requests allowed)
type NoAuth struct{}

// NewNoAuth creates a new NoAuth authenticator
func NewNoAuth() *NoAuth {
	return &NoAuth{}
}

// Authenticate always returns a dummy user (no authentication)
func (a *NoAuth) Authenticate(r *http.Request) (*User, error) {
	return &User{Username: "anonymous"}, nil
}

// Middleware returns a no-op middleware (passes all requests through)
func (a *NoAuth) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}
}
