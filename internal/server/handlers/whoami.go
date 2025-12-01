package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/criteo/command-launcher-registry/internal/auth"
)

// WhoamiHandler handles whoami requests
type WhoamiHandler struct {
	authenticator auth.Authenticator
	logger        *slog.Logger
}

// NewWhoamiHandler creates a new whoami handler
func NewWhoamiHandler(authenticator auth.Authenticator, logger *slog.Logger) *WhoamiHandler {
	return &WhoamiHandler{
		authenticator: authenticator,
		logger:        logger,
	}
}

// WhoamiResponse represents the whoami response
type WhoamiResponse struct {
	Username string `json:"username"`
}

// GetWhoami handles GET /api/v1/whoami
// This endpoint requires authentication and returns the authenticated username
func (h *WhoamiHandler) GetWhoami(w http.ResponseWriter, r *http.Request) {
	// Authenticate the request
	user, err := h.authenticator.Authenticate(r)
	if err != nil {
		h.logger.Debug("Authentication failed for whoami", "error", err)
		w.Header().Set("WWW-Authenticate", `Basic realm="COLA Registry"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Return the authenticated username
	response := WhoamiResponse{
		Username: user.Username,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode whoami response", "error", err)
	}
}
