package handlers

import (
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

	// Match AuthRequired: only the proxy-managed identity header is trusted.
	username := c.GetHeader("Remote-User")

	if username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated: Remote-User header not found"})
		return
	}
	if !auth.ProxyUserAllowed(username) {
		c.JSON(http.StatusForbidden, gin.H{"error": "User is not authorized"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"username": username,
		},
	})
}
