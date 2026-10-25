package auth

import (
	"os"
	"strings"
)

// ProxyUserAllowed checks the explicit list of users accepted from the trusted
// reverse proxy. An empty configuration accepts no proxy identities.
func ProxyUserAllowed(username string) bool {
	if username == "" {
		return false
	}
	for _, allowed := range strings.Split(os.Getenv("AUTH_ALLOWED_USERS"), ",") {
		if allowed = strings.TrimSpace(allowed); allowed != "" && username == allowed {
			return true
		}
	}
	return false
}
