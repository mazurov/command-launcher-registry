package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mazurov/command-launcher-registry/pkg/types"
)

// RequireTeamMembership checks if user belongs to allowed teams (for write access)
// If allowedTeams is empty, all authenticated users are granted access (useful for testing with personal GitHub accounts)
func RequireTeamMembership(allowedTeams []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// If no teams specified, allow all authenticated users (dev/testing mode)
		if len(allowedTeams) == 0 {
			c.Next()
			return
		}

		userTeams, exists := c.Get("user_teams")
		if !exists {
			c.JSON(http.StatusForbidden, types.ErrorResponse{
				Error:   "forbidden",
				Message: "User teams not found in context",
			})
			c.Abort()
			return
		}

		teams, ok := userTeams.([]string)
		if !ok {
			c.JSON(http.StatusForbidden, types.ErrorResponse{
				Error:   "forbidden",
				Message: "Invalid user teams format",
			})
			c.Abort()
			return
		}

		// Check if user is in any allowed team
		if !hasTeamMembership(teams, allowedTeams) {
			c.JSON(http.StatusForbidden, types.ErrorResponse{
				Error:   "forbidden",
				Message: "User does not have required team membership",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func hasTeamMembership(userTeams, allowedTeams []string) bool {
	for _, userTeam := range userTeams {
		for _, allowedTeam := range allowedTeams {
			if userTeam == allowedTeam {
				return true
			}
		}
	}
	return false
}
