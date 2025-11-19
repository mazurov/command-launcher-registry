package middleware

import (
	"net/http"

	"github.com/mazurov/command-launcher-registry/pkg/types"
	"github.com/gin-gonic/gin"
)

// ErrorHandler handles panics and converts them to error responses
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.WithField("error", err).Error("Panic recovered")
				c.JSON(http.StatusInternalServerError, types.ErrorResponse{
					Error:   "internal_error",
					Message: "An internal error occurred",
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
