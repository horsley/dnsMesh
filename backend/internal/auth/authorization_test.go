package auth

import "testing"

func TestProxyUserAllowed(t *testing.T) {
	for _, tc := range []struct {
		name, config, user string
		want               bool
	}{
		{"empty configuration denies", "", "horsley", false},
		{"empty identity denies", "horsley", "", false},
		{"owner allowed", "horsley", "horsley", true},
		{"mx denied", "horsley", "mx", false},
		{"sep denied", "horsley", "sep", false},
		{"exact match only", "horsley", "horsley-admin", false},
		{"case sensitive", "horsley", "Horsley", false},
		{"multiple explicit users", " horsley, second-user , ", "second-user", true},
		{"no wildcard interpretation", "*", "horsley", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("AUTH_ALLOWED_USERS", tc.config)
			if got := ProxyUserAllowed(tc.user); got != tc.want {
				t.Fatalf("ProxyUserAllowed(%q) = %t, want %t", tc.user, got, tc.want)
			}
		})
	}
}
