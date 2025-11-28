package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// User represents an authenticated user
type User struct {
	ID        string    `json:"id"`         // Provider-specific user ID
	Username  string    `json:"username"`   // Login name
	Email     string    `json:"email"`      // Primary email
	Name      string    `json:"name"`       // Display name
	AvatarURL string    `json:"avatar_url"` // Profile picture URL
	Teams     []string  `json:"teams"`      // Team/group slugs
	Provider  string    `json:"provider"`   // "github", "ldap", etc.
	CreatedAt time.Time `json:"created_at"`
}

// JWTClaims represents JWT token payload
type JWTClaims struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Teams    []string `json:"teams"`
	Provider string   `json:"provider"`
	jwt.RegisteredClaims
}
