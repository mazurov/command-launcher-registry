package auth

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v3"
)

// UserConfig represents a user in the users.yaml file
type UserConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"` // bcrypt hash
}

// UsersFile represents the structure of users.yaml
type UsersFile struct {
	Users []UserConfig `yaml:"users"`
}

// BasicAuth implements HTTP Basic Authentication
type BasicAuth struct {
	users  map[string]string // username -> bcrypt hash
	logger *slog.Logger
}

// NewBasicAuth creates a new BasicAuth authenticator
func NewBasicAuth(usersFile string, logger *slog.Logger) (*BasicAuth, error) {
	// Read users file
	data, err := os.ReadFile(usersFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read users file: %w", err)
	}

	// Parse YAML
	var usersFileData UsersFile
	if err := yaml.Unmarshal(data, &usersFileData); err != nil {
		return nil, fmt.Errorf("failed to parse users file (invalid YAML syntax): %w", err)
	}

	// Build username -> password hash map
	users := make(map[string]string)
	for _, user := range usersFileData.Users {
		users[user.Username] = user.Password
	}

	logger.Info("Basic auth initialized",
		"users_file", usersFile,
		"user_count", len(users))

	return &BasicAuth{
		users:  users,
		logger: logger,
	}, nil
}

// Authenticate validates HTTP Basic Auth credentials
func (a *BasicAuth) Authenticate(r *http.Request) (*User, error) {
	username, password, ok := r.BasicAuth()
	if !ok {
		return nil, fmt.Errorf("missing basic auth credentials")
	}

	// Check if user exists
	hashedPassword, exists := a.users[username]
	if !exists {
		a.logger.Warn("Authentication failed: user not found",
			"username", username,
			"source_ip", r.RemoteAddr)
		return nil, fmt.Errorf("invalid credentials")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		a.logger.Warn("Authentication failed: invalid password",
			"username", username,
			"source_ip", r.RemoteAddr)
		return nil, fmt.Errorf("invalid credentials")
	}

	a.logger.Debug("Authentication successful",
		"username", username,
		"source_ip", r.RemoteAddr)

	return &User{Username: username}, nil
}

// Middleware returns HTTP Basic Auth middleware
func (a *BasicAuth) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Authenticate request
			user, err := a.Authenticate(r)
			if err != nil {
				w.Header().Set("WWW-Authenticate", `Basic realm="COLA Registry"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Store user in context (if needed in future)
			_ = user

			next.ServeHTTP(w, r)
		})
	}
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
