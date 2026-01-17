package auth

import (
	"os"
	"strings"
)

// BypassUser returns the local development user and true when auth bypass is enabled.
func BypassUser() (string, bool) {
	if !isTruthy(os.Getenv("AUTH_BYPASS")) {
		return "", false
	}

	username := strings.TrimSpace(os.Getenv("AUTH_BYPASS_USER"))
	if username == "" {
		username = "local-dev"
	}

	return username, true
}

func isTruthy(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}
