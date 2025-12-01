package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/criteo/command-launcher-registry/internal/auth"
)

func TestWhoamiHandler_GetWhoami(t *testing.T) {
	logger := slog.Default()

	tests := []struct {
		name           string
		authType       string
		username       string
		password       string
		expectStatus   int
		expectUsername string
	}{
		{
			name:           "successful authentication",
			authType:       "basic",
			username:       "testuser",
			password:       "testpass",
			expectStatus:   http.StatusOK,
			expectUsername: "testuser",
		},
		{
			name:         "no authentication",
			authType:     "basic",
			expectStatus: http.StatusUnauthorized,
		},
		{
			name:         "invalid credentials",
			authType:     "basic",
			username:     "wronguser",
			password:     "wrongpass",
			expectStatus: http.StatusUnauthorized,
		},
		{
			name:           "no auth mode - anyone can authenticate",
			authType:       "none",
			expectStatus:   http.StatusOK,
			expectUsername: "anonymous",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create authenticator based on test case
			var authenticator auth.Authenticator
			if tt.authType == "none" {
				authenticator = auth.NewNoAuth()
			} else {
				// For basic auth tests, we'll use a mock that accepts testuser:testpass
				authenticator = &mockAuthenticator{
					validUsername: "testuser",
					validPassword: "testpass",
				}
			}

			handler := NewWhoamiHandler(authenticator, logger)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/api/v1/whoami", nil)
			if tt.username != "" && tt.password != "" {
				req.SetBasicAuth(tt.username, tt.password)
			}

			// Record response
			rr := httptest.NewRecorder()

			// Call handler
			handler.GetWhoami(rr, req)

			// Check status code
			if rr.Code != tt.expectStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, tt.expectStatus)
			}

			// Check response body for successful auth
			if tt.expectStatus == http.StatusOK {
				var response WhoamiResponse
				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if response.Username != tt.expectUsername {
					t.Errorf("handler returned wrong username: got %v want %v", response.Username, tt.expectUsername)
				}
			}

			// Check WWW-Authenticate header for 401
			if tt.expectStatus == http.StatusUnauthorized {
				wwwAuth := rr.Header().Get("WWW-Authenticate")
				if wwwAuth != `Basic realm="COLA Registry"` {
					t.Errorf("handler returned wrong WWW-Authenticate header: got %v", wwwAuth)
				}
			}
		})
	}
}

// mockAuthenticator is a simple mock for testing
type mockAuthenticator struct {
	validUsername string
	validPassword string
}

func (m *mockAuthenticator) Authenticate(r *http.Request) (*auth.User, error) {
	username, password, ok := r.BasicAuth()
	if !ok || username != m.validUsername || password != m.validPassword {
		return nil, fmt.Errorf("invalid credentials")
	}
	return &auth.User{Username: username}, nil
}

func (m *mockAuthenticator) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return next
	}
}
