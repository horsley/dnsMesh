package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthRequired is a middleware that checks Remote-User header for authentication
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try multiple common header names for reverse proxy authentication
		// Common headers: Remote-User, remote-user, X-Remote-User, X-Forwarded-User
		username := c.GetHeader("Remote-User")

		if username == "" {
			username = c.GetHeader("remote-user")
		}
		if username == "" {
			username = c.GetHeader("X-Remote-User")
		}
		if username == "" {
			username = c.GetHeader("X-Forwarded-User")
		}

		// Log all authentication-related headers for debugging
		log.Printf("Auth check - Remote-User: '%s', remote-user: '%s', X-Remote-User: '%s', X-Forwarded-User: '%s'",
			c.GetHeader("Remote-User"),
			c.GetHeader("remote-user"),
			c.GetHeader("X-Remote-User"),
			c.GetHeader("X-Forwarded-User"))

		if username == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required: Remote-User header not found",
			})
			c.Abort()
			return
		}

		// Set username in context
		c.Set("username", username)
		c.Next()
	}
}
