package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestRefreshRotatesStoredTokenFromFixture(t *testing.T) {
	tokenFixture, err := os.ReadFile("../../whoop/testdata/token.json")
	if err != nil {
		t.Fatal(err)
	}
	var gotRefresh string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		gotRefresh = r.Form.Get("refresh_token")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(tokenFixture)
	}))
	defer server.Close()
	store := FileStore{Path: t.TempDir() + "/credentials.json"}
	if err := store.Set(context.Background(), "refresh_token", "refresh-old"); err != nil {
		t.Fatal(err)
	}
	oauth, err := New(Config{ClientID: "client", ClientSecret: "secret", TokenURL: server.URL, Store: store})
	if err != nil {
		t.Fatal(err)
	}
	token, err := oauth.Refresh(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if token.AccessToken != "access-new" || gotRefresh != "refresh-old" {
		t.Fatalf("unexpected token request: %#v refresh=%q", token, gotRefresh)
	}
	stored, err := store.Get(context.Background(), "refresh_token")
	if err != nil || stored != "refresh-new" {
		t.Fatalf("rotated token not stored: %q %v", stored, err)
	}
}

func TestTokenErrorsRedactSecrets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_grant","error_description":"refresh_token=do-not-leak client_secret=also-secret"}`))
	}))
	defer server.Close()
	store := FileStore{Path: t.TempDir() + "/credentials.json"}
	_ = store.Set(context.Background(), "refresh_token", "refresh-old")
	oauth, _ := New(Config{ClientID: "client", ClientSecret: "secret", TokenURL: server.URL, Store: store})
	_, err := oauth.Refresh(context.Background())
	if err == nil || strings.Contains(err.Error(), "do-not-leak") || strings.Contains(err.Error(), "also-secret") {
		t.Fatalf("expected redacted token error: %v", err)
	}
}

func TestFileStoreMissingAndJSONShape(t *testing.T) {
	store := FileStore{Path: t.TempDir() + "/credentials.json"}
	_, err := store.Get(context.Background(), "refresh_token")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := store.Set(context.Background(), "client_id", "abc"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(store.Path)
	if err != nil {
		t.Fatal(err)
	}
	var values map[string]string
	if json.Unmarshal(data, &values) != nil || values["client_id"] != "abc" {
		t.Fatalf("unexpected credential JSON: %s", data)
	}
}

func TestFileStoreRejectsMalformedDataWithoutOverwriting(t *testing.T) {
	path := t.TempDir() + "/credentials.json"
	if err := os.WriteFile(path, []byte(`{"refresh_token":`), 0600); err != nil {
		t.Fatal(err)
	}
	store := FileStore{Path: path}
	err := store.Set(context.Background(), "client_id", "new-client")
	if err == nil || !strings.Contains(err.Error(), "decode credential file") {
		t.Fatalf("expected malformed credential error, got %v", err)
	}
	data, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(data) != `{"refresh_token":` {
		t.Fatalf("malformed credential file was overwritten: %q", data)
	}
}
