package handlers

import (
	"log"
	"net/http"

	"dnsmesh/internal/auth"

	"github.com/gin-gonic/gin"
)

// GetCurrentUser returns the current authenticated user from Remote-User header
func GetCurrentUser(c *gin.Context) {
	if username, ok := auth.BypassUser(); ok {
		c.JSON(http.StatusOK, gin.H{
			"user": gin.H{
				"username": username,
			},
		})
		return
	}

	// Try multiple common header names for reverse proxy authentication
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

	// Log for debugging
	log.Printf("GetCurrentUser - Remote-User: '%s', remote-user: '%s', X-Remote-User: '%s', X-Forwarded-User: '%s', Final: '%s'",
		c.GetHeader("Remote-User"),
		c.GetHeader("remote-user"),
		c.GetHeader("X-Remote-User"),
		c.GetHeader("X-Forwarded-User"),
		username)

	if username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated: Remote-User header not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"username": username,
		},
	})
}
