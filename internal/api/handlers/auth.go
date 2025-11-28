package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mazurov/command-launcher-registry/internal/auth"
	"github.com/mazurov/command-launcher-registry/internal/auth/provider"
	"github.com/mazurov/command-launcher-registry/pkg/types"
)

// stateInfo stores OAuth state information
type stateInfo struct {
	ExpiresAt    time.Time
	CallbackPort string // For CLI interactive mode
}

// AuthHandler handles authentication requests
type AuthHandler struct {
	provider          provider.AuthProvider
	config            *auth.Config
	stateStore        map[string]*stateInfo // state -> info for OAuth state validation
	cleanupDone       chan struct{}
	DeviceFlowHandler *DeviceFlowHandler // For device flow (set after creation)
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authProvider provider.AuthProvider, cfg *auth.Config) *AuthHandler {
	h := &AuthHandler{
		provider:    authProvider,
		config:      cfg,
		stateStore:  make(map[string]*stateInfo),
		cleanupDone: make(chan struct{}),
	}

	// Start cleanup goroutine for expired states (every 5 minutes)
	go h.cleanupExpiredStates()

	return h
}

// HandleLogin initiates OAuth flow
func (h *AuthHandler) HandleLogin(c *gin.Context) {
	// Get callback_port parameter (for CLI interactive mode)
	callbackPort := c.Query("callback_port")

	// Generate random state for CSRF protection
	state, err := h.generateState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to generate state",
		})
		return
	}

	// Store state with timestamp and callback port (valid for 10 minutes)
	h.stateStore[state] = &stateInfo{
		ExpiresAt:    time.Now().Add(10 * time.Minute),
		CallbackPort: callbackPort,
	}

	// Get authorization URL
	authURL := h.provider.GetAuthURL(state)

	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
		"state":    state,
	})
}

// HandleCallback processes OAuth callback
func (h *AuthHandler) HandleCallback(c *gin.Context) {
	// Get code and state from query parameters
	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "invalid_request",
			Message: "Missing code or state parameter",
		})
		return
	}

	// Validate state
	info, exists := h.stateStore[state]
	if !exists {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "invalid_state",
			Message: "Invalid or expired state parameter",
		})
		return
	}

	// Check if state is expired
	if time.Now().After(info.ExpiresAt) {
		delete(h.stateStore, state)
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "expired_state",
			Message: "State parameter has expired",
		})
		return
	}

	// Get callback port from stored state info (not from query param)
	callbackPort := info.CallbackPort

	// Remove used state
	delete(h.stateStore, state)

	// Exchange code for user info
	user, err := h.provider.HandleCallback(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Error:   "authentication_failed",
			Message: fmt.Sprintf("Failed to authenticate: %v", err),
		})
		return
	}

	// Check if this is a device flow callback (state starts with "device_")
	if strings.HasPrefix(state, "device_") && h.DeviceFlowHandler != nil {
		h.DeviceFlowHandler.HandleDeviceCallback(c, user, state)
		return
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user, h.config.JWTSecret, h.config.TokenExpiry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   "token_generation_failed",
			Message: "Failed to generate authentication token",
		})
		return
	}

	// If callback_port is specified, redirect to CLI callback server
	// This is for interactive CLI mode (deprecated, use device flow instead)
	if callbackPort != "" {
		redirectURL := fmt.Sprintf("http://localhost:%s/callback?token=%s", callbackPort, token)
		c.Redirect(http.StatusFound, redirectURL)
		return
	}

	// Otherwise, return JSON response (browser/API mode)
	c.JSON(http.StatusOK, gin.H{
		"token":      token,
		"token_type": "Bearer",
		"expires_in": h.config.TokenExpiry * 3600, // Convert hours to seconds
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"name":     user.Name,
			"teams":    user.Teams,
			"provider": user.Provider,
		},
	})
}

// GetCurrentUser returns current authenticated user info
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	// User info is already in context from AuthMiddleware
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")
	userEmail, _ := c.Get("user_email")
	userTeams, _ := c.Get("user_teams")
	userProvider, _ := c.Get("user_provider")

	// Convert teams to []string if needed
	teams, ok := userTeams.([]string)
	if !ok {
		teams = []string{}
	}

	// Convert email to string if needed
	email, ok := userEmail.(string)
	if !ok {
		email = ""
	}

	c.JSON(http.StatusOK, gin.H{
		"id":       userID,
		"username": username,
		"email":    email,
		"teams":    teams,
		"provider": userProvider,
	})
}

// ExchangeGitHubPAT exchanges GitHub Personal Access Token for JWT (for CI/CD)
// This allows non-interactive authentication using GitHub PAT
func (h *AuthHandler) ExchangeGitHubPAT(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "invalid_request",
			Message: "Missing or invalid token in request body",
		})
		return
	}

	// This endpoint is only for GitHub provider
	githubProvider, ok := h.provider.(interface {
		ValidatePAT(ctx context.Context, token string) (*auth.User, error)
	})
	if !ok {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "unsupported_provider",
			Message: "PAT exchange is only supported for GitHub provider",
		})
		return
	}

	// Validate PAT and get user info
	user, err := githubProvider.ValidatePAT(c.Request.Context(), req.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Error:   "invalid_token",
			Message: fmt.Sprintf("Failed to validate GitHub PAT: %v", err),
		})
		return
	}

	// Generate JWT token
	jwtToken, err := auth.GenerateToken(user, h.config.JWTSecret, h.config.TokenExpiry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   "token_generation_failed",
			Message: "Failed to generate JWT token",
		})
		return
	}

	// Return JWT token
	c.JSON(http.StatusOK, gin.H{
		"access_token": jwtToken,
		"token_type":   "Bearer",
		"expires_in":   h.config.TokenExpiry * 3600,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"teams":    user.Teams,
			"provider": user.Provider,
		},
	})
}

// generateState generates cryptographically secure random state
func (h *AuthHandler) generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// cleanupExpiredStates removes expired states from store
func (h *AuthHandler) cleanupExpiredStates() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			for state, info := range h.stateStore {
				if now.After(info.ExpiresAt) {
					delete(h.stateStore, state)
				}
			}
		case <-h.cleanupDone:
			return
		}
	}
}

// Shutdown stops the cleanup goroutine
func (h *AuthHandler) Shutdown() {
	close(h.cleanupDone)
}
