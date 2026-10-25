package middleware

import (
	"crypto/subtle"
	"net/http"
	"os"

	"dnsmesh/internal/auth"

	"github.com/gin-gonic/gin"
)

// AuthRequired is a middleware that checks Remote-User header for authentication
// Also supports API Key authentication via X-API-Key header
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for API Key first
		apiKey := c.GetHeader("X-API-Key")
		if apiKey != "" {
			expectedKey := os.Getenv("API_TOKEN")
			if expectedKey != "" && subtle.ConstantTimeCompare([]byte(apiKey), []byte(expectedKey)) == 1 {
				c.Set("username", "api-user")
				c.Next()
				return
			}
			// A supplied credential must not silently fall back to a browser identity.
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
			return
		}

		// Check bypass auth (development mode)
		if username, ok := auth.BypassUser(); ok {
			c.Set("username", username)
			c.Next()
			return
		}

		// Only trust the identity header populated and cleared by our reverse proxy.
		// Header lookup is case-insensitive; alternate identity names are not trusted.
		username := c.GetHeader("Remote-User")

		if username == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required: Remote-User header not found",
			})
			c.Abort()
			return
		}
		if !auth.ProxyUserAllowed(username) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "User is not authorized"})
			return
		}

		// Set username in context
		c.Set("username", username)
		c.Next()
	}
}
