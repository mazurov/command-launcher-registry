package middleware

import (
	"net/http"

	"github.com/criteo/command-launcher-registry/internal/auth"
)

// RequireAuth returns middleware that requires authentication for write operations
// Read operations (GET) are allowed without authentication
func RequireAuth(authenticator auth.Authenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if this is a write operation
			if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodDelete {
				// Require authentication
				if _, err := authenticator.Authenticate(r); err != nil {
					w.Header().Set("WWW-Authenticate", `Basic realm="COLA Registry"`)
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
