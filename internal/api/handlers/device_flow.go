package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mazurov/command-launcher-registry/internal/auth"
	"github.com/mazurov/command-launcher-registry/pkg/types"
)

// DeviceCode represents a device authorization request
type DeviceCode struct {
	DeviceCode      string    `json:"device_code"`      // Code CLI uses to poll
	UserCode        string    `json:"user_code"`        // Code user enters in browser
	VerificationURI string    `json:"verification_uri"` // URL user opens
	ExpiresAt       time.Time `json:"expires_at"`       // When codes expire
	Interval        int       `json:"interval"`         // Polling interval in seconds

	// Internal state
	Authorized bool     `json:"-"` // Whether user has authorized
	UserID     string   `json:"-"` // User who authorized (from OAuth)
	Username   string   `json:"-"`
	Email      string   `json:"-"`
	Teams      []string `json:"-"`
	Provider   string   `json:"-"`
}

// DeviceFlowHandler manages device authorization flow
type DeviceFlowHandler struct {
	authHandler *AuthHandler
	codes       map[string]*DeviceCode // device_code -> DeviceCode
	userCodes   map[string]string      // user_code -> device_code
	mu          sync.RWMutex
	cleanupDone chan struct{}
}

// NewDeviceFlowHandler creates a new device flow handler
func NewDeviceFlowHandler(authHandler *AuthHandler) *DeviceFlowHandler {
	h := &DeviceFlowHandler{
		authHandler: authHandler,
		codes:       make(map[string]*DeviceCode),
		userCodes:   make(map[string]string),
		cleanupDone: make(chan struct{}),
	}

	// Start cleanup goroutine for expired codes
	go h.cleanupExpiredCodes()

	return h
}

// HandleDeviceCode generates device and user codes (CLI calls this)
func (h *DeviceFlowHandler) HandleDeviceCode(c *gin.Context) {
	// Generate device code (long, for CLI polling)
	deviceCode, err := h.generateRandomCode(32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to generate device code",
		})
		return
	}

	// Generate user code (short, human-readable)
	userCode, err := h.generateUserCode()
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to generate user code",
		})
		return
	}

	// Create device code entry (expires in 15 minutes)
	code := &DeviceCode{
		DeviceCode:      deviceCode,
		UserCode:        userCode,
		VerificationURI: fmt.Sprintf("%s/auth/device", getBaseURL(c)),
		ExpiresAt:       time.Now().Add(15 * time.Minute),
		Interval:        5, // Poll every 5 seconds
		Authorized:      false,
	}

	// Store codes
	h.mu.Lock()
	h.codes[deviceCode] = code
	h.userCodes[userCode] = deviceCode
	h.mu.Unlock()

	// Return response to CLI
	c.JSON(http.StatusOK, gin.H{
		"device_code":      code.DeviceCode,
		"user_code":        code.UserCode,
		"verification_uri": code.VerificationURI,
		"expires_in":       int(time.Until(code.ExpiresAt).Seconds()),
		"interval":         code.Interval,
	})
}

// HandleDeviceAuthorize shows authorization page (user opens in browser)
func (h *DeviceFlowHandler) HandleDeviceAuthorize(c *gin.Context) {
	userCode := c.Query("user_code")

	// If this is the initial page load (no user_code yet), show the form
	if userCode == "" {
		h.showAuthorizationForm(c, "", "")
		return
	}

	// Validate user code
	h.mu.RLock()
	deviceCode, exists := h.userCodes[userCode]
	h.mu.RUnlock()

	if !exists {
		h.showAuthorizationForm(c, userCode, "Invalid or expired code")
		return
	}

	h.mu.RLock()
	code := h.codes[deviceCode]
	h.mu.RUnlock()

	if code == nil || time.Now().After(code.ExpiresAt) {
		h.showAuthorizationForm(c, userCode, "Code has expired")
		return
	}

	// Code is valid, redirect to GitHub OAuth with special state
	state := "device_" + deviceCode

	// Store state for OAuth callback
	h.authHandler.stateStore[state] = &stateInfo{
		ExpiresAt:    time.Now().Add(10 * time.Minute),
		CallbackPort: "", // No callback port for device flow
	}

	authURL := h.authHandler.provider.GetAuthURL(state)
	c.Redirect(http.StatusFound, authURL)
}

// HandleDeviceCallback processes OAuth callback for device flow
func (h *DeviceFlowHandler) HandleDeviceCallback(c *gin.Context, user *auth.User, state string) {
	// Extract device_code from state
	if !strings.HasPrefix(state, "device_") {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "invalid_state",
			Message: "Invalid state for device flow",
		})
		return
	}

	deviceCode := strings.TrimPrefix(state, "device_")

	// Find device code entry
	h.mu.Lock()
	code, exists := h.codes[deviceCode]
	if !exists || code == nil {
		h.mu.Unlock()
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "invalid_device_code",
			Message: "Device code not found or expired",
		})
		return
	}

	// Mark as authorized and store user info
	code.Authorized = true
	code.UserID = user.ID
	code.Username = user.Username
	code.Email = user.Email
	code.Teams = user.Teams
	code.Provider = user.Provider
	h.mu.Unlock()

	// Show success page to user
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, `
<!DOCTYPE html>
<html>
<head>
    <title>Authorization Successful</title>
    <style>
        body { font-family: Arial, sans-serif; text-align: center; padding: 50px; }
        .success { color: #28a745; font-size: 24px; margin: 20px 0; }
        .info { color: #666; margin: 10px 0; }
    </style>
</head>
<body>
    <h1 class="success">Authorization Successful!</h1>
    <p class="info">You have successfully authorized the device.</p>
    <p class="info">You can close this window and return to the CLI.</p>
</body>
</html>
`)
}

// HandleDeviceToken polls for token (CLI calls this repeatedly)
func (h *DeviceFlowHandler) HandleDeviceToken(c *gin.Context) {
	var req struct {
		DeviceCode string `json:"device_code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "invalid_request",
			Message: "Missing device_code parameter",
		})
		return
	}

	// Find device code
	h.mu.RLock()
	code, exists := h.codes[req.DeviceCode]
	h.mu.RUnlock()

	if !exists || code == nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "invalid_grant",
			Message: "Invalid or expired device code",
		})
		return
	}

	// Check expiration
	if time.Now().After(code.ExpiresAt) {
		h.mu.Lock()
		delete(h.codes, req.DeviceCode)
		delete(h.userCodes, code.UserCode)
		h.mu.Unlock()

		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "expired_token",
			Message: "Device code has expired",
		})
		return
	}

	// Check if authorized
	if !code.Authorized {
		// Still waiting for user authorization
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error:   "authorization_pending",
			Message: "User has not yet authorized the device",
		})
		return
	}

	// Generate JWT token
	user := &auth.User{
		ID:       code.UserID,
		Username: code.Username,
		Email:    code.Email,
		Teams:    code.Teams,
		Provider: code.Provider,
	}

	token, err := auth.GenerateToken(user, h.authHandler.config.JWTSecret, h.authHandler.config.TokenExpiry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error:   "token_generation_failed",
			Message: "Failed to generate authentication token",
		})
		return
	}

	// Clean up device code (one-time use)
	h.mu.Lock()
	delete(h.codes, req.DeviceCode)
	delete(h.userCodes, code.UserCode)
	h.mu.Unlock()

	// Return token
	c.JSON(http.StatusOK, gin.H{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_in":   h.authHandler.config.TokenExpiry * 3600,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"teams":    user.Teams,
			"provider": user.Provider,
		},
	})
}

// showAuthorizationForm displays the code entry form
func (h *DeviceFlowHandler) showAuthorizationForm(c *gin.Context, userCode, errorMsg string) {
	errorHTML := ""
	if errorMsg != "" {
		errorHTML = fmt.Sprintf(`<p class="error">%s</p>`, errorMsg)
	}

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>Device Authorization</title>
    <style>
        body { font-family: Arial, sans-serif; text-align: center; padding: 50px; background: #f5f5f5; }
        .container { max-width: 400px; margin: 0 auto; background: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        h1 { color: #333; margin-bottom: 20px; }
        .info { color: #666; margin: 15px 0; }
        input { width: 100%%; padding: 12px; font-size: 18px; text-align: center; border: 2px solid #ddd; border-radius: 4px; margin: 15px 0; text-transform: uppercase; letter-spacing: 2px; }
        button { width: 100%%; padding: 12px; font-size: 16px; background: #28a745; color: white; border: none; border-radius: 4px; cursor: pointer; }
        button:hover { background: #218838; }
        .error { color: #dc3545; margin: 15px 0; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Device Authorization</h1>
        <p class="info">Enter the code shown in your CLI:</p>
        %s
        <form method="GET" action="/auth/device">
            <input type="text" name="user_code" placeholder="XXXX-XXXX" value="%s" required autofocus>
            <button type="submit">Continue</button>
        </form>
    </div>
</body>
</html>
`, errorHTML, userCode))
}

// generateRandomCode generates a random code
func (h *DeviceFlowHandler) generateRandomCode(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// generateUserCode generates a human-readable code (e.g., "ABCD-1234")
func (h *DeviceFlowHandler) generateUserCode() (string, error) {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // Avoid similar looking chars
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	code := make([]byte, 8)
	for i := range b {
		code[i] = charset[int(b[i])%len(charset)]
	}

	// Format as XXXX-XXXX
	return fmt.Sprintf("%s-%s", string(code[:4]), string(code[4:])), nil
}

// cleanupExpiredCodes removes expired device codes
func (h *DeviceFlowHandler) cleanupExpiredCodes() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			h.mu.Lock()
			for deviceCode, code := range h.codes {
				if now.After(code.ExpiresAt) {
					delete(h.codes, deviceCode)
					delete(h.userCodes, code.UserCode)
				}
			}
			h.mu.Unlock()
		case <-h.cleanupDone:
			return
		}
	}
}

// Shutdown stops the cleanup goroutine
func (h *DeviceFlowHandler) Shutdown() {
	close(h.cleanupDone)
}

// getBaseURL extracts base URL from request
func getBaseURL(c *gin.Context) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s", scheme, c.Request.Host)
}
