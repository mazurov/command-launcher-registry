package middleware

import (
	"net/http"
	"strings"
)

// CORS returns middleware that handles CORS for index.json endpoint only
func CORS() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if this is an index.json endpoint
			if strings.HasSuffix(r.URL.Path, "/index.json") {
				// Set CORS headers
				w.Header().Set("Access-Control-Allow-Origin", "*")
				w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
				w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

				// Handle OPTIONS preflight
				if r.Method == http.MethodOptions {
					w.WriteHeader(http.StatusOK)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
