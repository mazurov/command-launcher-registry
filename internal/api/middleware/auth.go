package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mazurov/command-launcher-registry/internal/auth"
	"github.com/mazurov/command-launcher-registry/pkg/types"
)

// AuthMiddleware validates JWT token and populates user context
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, types.ErrorResponse{
				Error:   "unauthorized",
				Message: "Missing authorization header",
			})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, types.ErrorResponse{
				Error:   "unauthorized",
				Message: "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := auth.ValidateToken(tokenString, jwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, types.ErrorResponse{
				Error:   "unauthorized",
				Message: "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Store user info in context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("user_email", claims.Email)
		c.Set("user_teams", claims.Teams)
		c.Set("user_provider", claims.Provider)

		c.Next()
	}
}

// JWTAuth is deprecated, use AuthMiddleware instead
// Kept for backward compatibility
func JWTAuth(jwtSecret string) gin.HandlerFunc {
	return AuthMiddleware(jwtSecret)
}

// APIKeyAuth validates API keys (simple implementation)
func APIKeyAuth(validKeys map[string]bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, types.ErrorResponse{
				Error:   "unauthorized",
				Message: "Missing API key",
			})
			c.Abort()
			return
		}

		if !validKeys[apiKey] {
			c.JSON(http.StatusUnauthorized, types.ErrorResponse{
				Error:   "unauthorized",
				Message: "Invalid API key",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
