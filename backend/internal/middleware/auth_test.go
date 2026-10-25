package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAuthRequiredIdentityBoundary(t *testing.T) {
	t.Setenv("AUTH_BYPASS", "false")
	t.Setenv("API_TOKEN", "test-only-api-key")
	t.Setenv("AUTH_ALLOWED_USERS", "horsley")
	for _, tc := range []struct {
		name, header, value string
		wantStatus          int
	}{
		{"anonymous", "", "", http.StatusUnauthorized},
		{"valid API key", "X-API-Key", "test-only-api-key", http.StatusOK},
		{"invalid API key", "X-API-Key", "invalid", http.StatusUnauthorized},
		{"proxy user", "Remote-User", "horsley", http.StatusOK},
		{"case insensitive proxy user", "remote-user", "horsley", http.StatusOK},
		{"other authenticated user mx", "Remote-User", "mx", http.StatusForbidden},
		{"other authenticated user sep", "Remote-User", "sep", http.StatusForbidden},
		{"untrusted alternate user", "X-Remote-User", "forged", http.StatusUnauthorized},
		{"untrusted forwarded user", "X-Forwarded-User", "forged", http.StatusUnauthorized},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()
			r.GET("/protected", AuthRequired(), func(c *gin.Context) { c.Status(http.StatusOK) })
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tc.header != "" {
				req.Header.Set(tc.header, tc.value)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", w.Code, tc.wantStatus)
			}
		})
	}
}

func TestInvalidAPIKeyDoesNotFallBackToAllowedUser(t *testing.T) {
	t.Setenv("AUTH_BYPASS", "false")
	t.Setenv("API_TOKEN", "test-only-api-key")
	t.Setenv("AUTH_ALLOWED_USERS", "horsley")
	r := gin.New()
	r.GET("/protected", AuthRequired(), func(c *gin.Context) { c.Status(http.StatusOK) })
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("X-API-Key", "invalid")
	req.Header.Set("Remote-User", "horsley")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}
