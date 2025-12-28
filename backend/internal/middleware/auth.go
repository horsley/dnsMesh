package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthRequired is a middleware that checks Remote-User header for authentication
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get username from Remote-User header (set by reverse proxy)
		username := c.GetHeader("Remote-User")

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
