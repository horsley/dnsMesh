package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGetCurrentUserRejectsUntrustedIdentityAliases(t *testing.T) {
	t.Setenv("AUTH_BYPASS", "false")
	t.Setenv("AUTH_ALLOWED_USERS", "horsley")
	for _, tc := range []struct {
		header string
		want   int
	}{
		{"Remote-User", http.StatusOK},
		{"remote-user", http.StatusOK},
		{"X-Remote-User", http.StatusUnauthorized},
		{"X-Forwarded-User", http.StatusUnauthorized},
	} {
		t.Run(tc.header, func(t *testing.T) {
			r := gin.New()
			r.GET("/api/auth/user", GetCurrentUser)
			req := httptest.NewRequest(http.MethodGet, "/api/auth/user", nil)
			req.Header.Set(tc.header, "horsley")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tc.want {
				t.Fatalf("status = %d, want %d", w.Code, tc.want)
			}
		})
	}
}

func TestGetCurrentUserRejectsOtherAuthenticatedUsers(t *testing.T) {
	t.Setenv("AUTH_BYPASS", "false")
	t.Setenv("AUTH_ALLOWED_USERS", "horsley")
	for _, username := range []string{"mx", "sep"} {
		t.Run(username, func(t *testing.T) {
			r := gin.New()
			r.GET("/api/auth/user", GetCurrentUser)
			req := httptest.NewRequest(http.MethodGet, "/api/auth/user", nil)
			req.Header.Set("Remote-User", username)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != http.StatusForbidden {
				t.Fatalf("status = %d, want %d", w.Code, http.StatusForbidden)
			}
		})
	}
}
