package auth

import (
	"net/url"
	"strings"
	"testing"
)

func TestAuthorizationURLCarriesStateAndLeastPrivilegeScopes(t *testing.T) {
	oauth, err := New(Config{ClientID: "client", ClientSecret: "secret", Store: FileStore{Path: t.TempDir() + "/credentials.json"}})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := oauth.AuthorizationURL("csrf-state")
	if err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	if q.Get("state") != "csrf-state" || q.Get("response_type") != "code" || q.Get("redirect_uri") != DefaultRedirectURL {
		t.Fatalf("unexpected OAuth URL: %s", raw)
	}
	for _, required := range []string{"read:profile", "read:body_measurement", "read:cycles", "read:recovery", "read:sleep", "read:workout", "offline"} {
		if !containsScope(q.Get("scope"), required) {
			t.Fatalf("scope %q missing from %q", required, q.Get("scope"))
		}
	}
}

func containsScope(scopes, required string) bool {
	for _, scope := range strings.Fields(scopes) {
		if scope == required {
			return true
		}
	}
	return false
}
