package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetCurrentUser returns the current authenticated user from Remote-User header
func GetCurrentUser(c *gin.Context) {
	// Get username from context (set by middleware)
	username, exists := c.Get("username")

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"username": username,
		},
	})
}
