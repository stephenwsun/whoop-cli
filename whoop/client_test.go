package whoop

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestWorkoutsFollowNextTokenFixtures(t *testing.T) {
	var tokens []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokens = append(tokens, r.URL.Query().Get("next_token"))
		w.Header().Set("Content-Type", "application/json")
		if len(tokens) == 1 {
			_, _ = w.Write(fixture(t, "workouts-page1.json"))
			return
		}
		_, _ = w.Write(fixture(t, "workouts-page2.json"))
	}))
	defer server.Close()
	client, err := NewClient(ClientConfig{BaseURL: server.URL, AccessToken: "secret-access", MaxRetries: 0})
	if err != nil {
		t.Fatal(err)
	}
	workouts, err := client.Workouts(context.Background(), ListOptions{Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(workouts) != 2 || workouts[1].SportName != "muay-thai" {
		t.Fatalf("unexpected workouts: %#v", workouts)
	}
	if len(tokens) != 2 || tokens[0] != "" || tokens[1] != "page-2" {
		t.Fatalf("unexpected pagination tokens: %#v", tokens)
	}
}

func TestTransportRetriesRateLimitAndRedactsError(t *testing.T) {
	attempts := 0
	slept := time.Duration(0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write(fixture(t, "error.json"))
			return
		}
		_, _ = w.Write(fixture(t, "profile.json"))
	}))
	defer server.Close()
	client, err := NewClient(ClientConfig{BaseURL: server.URL, MaxRetries: 1, RetryBackoff: func(int) time.Duration { return time.Hour }, Sleep: func(_ context.Context, d time.Duration) error { slept = d; return nil }})
	if err != nil {
		t.Fatal(err)
	}
	profile, err := client.Profile(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if profile.UserID != "user-1" || attempts != 2 || slept != time.Second {
		t.Fatalf("retry behavior: profile=%#v attempts=%d slept=%s", profile, attempts, slept)
	}

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(fixture(t, "error.json"))
	}))
	defer server2.Close()
	client, _ = NewClient(ClientConfig{BaseURL: server2.URL, MaxRetries: 0})
	_, err = client.Profile(context.Background())
	if err == nil || strings.Contains(err.Error(), "do-not-leak") {
		t.Fatalf("expected redacted API error, got %v", err)
	}
}

func TestPaginationRejectsRepeatedToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"records":[],"next_token":"same"}`))
	}))
	defer server.Close()
	client, _ := NewClient(ClientConfig{BaseURL: server.URL, MaxRetries: 0})
	_, err := client.Workouts(context.Background(), ListOptions{})
	var paginationErr *PaginationError
	if !errors.As(err, &paginationErr) {
		t.Fatalf("expected pagination error, got %v", err)
	}
}

func TestWorkoutsPageDoesNotFetchNextPage(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if got := r.URL.Query().Get("next_token"); got != "cursor-1" {
			t.Errorf("next_token = %q, want cursor-1", got)
		}
		_, _ = w.Write(fixture(t, "workouts-page1.json"))
	}))
	defer server.Close()

	client, err := NewClient(ClientConfig{BaseURL: server.URL, MaxRetries: 0})
	if err != nil {
		t.Fatal(err)
	}
	page, err := client.WorkoutsPage(context.Background(), ListOptions{Limit: 1, NextToken: "cursor-1"})
	if err != nil {
		t.Fatal(err)
	}
	if requests != 1 || len(page.Records) != 1 || page.NextToken != "page-2" {
		t.Fatalf("unexpected page result: requests=%d page=%#v", requests, page)
	}
}
