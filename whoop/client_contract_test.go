package whoop

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResourcePagesDecodeDocumentedFixtures(t *testing.T) {
	paths := map[string]string{
		"/developer/v2/user/profile/basic":    "profile.json",
		"/developer/v2/user/measurement/body": "body-page.json",
		"/developer/v2/cycle":                 "cycle-page.json",
		"/developer/v2/recovery":              "recovery-page.json",
		"/developer/v2/activity/sleep":        "sleep-page.json",
		"/developer/v2/activity/workout":      "workouts-page1.json",
	}
	seen := map[string]int{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen[r.URL.Path]++
		if got := r.Header.Get("Authorization"); got != "Bearer contract-token" {
			t.Errorf("Authorization = %q, want Bearer contract-token", got)
		}
		if r.URL.Path == "/developer/v2/user/measurement/body" {
			query := r.URL.Query()
			if query.Get("limit") != "1" || query.Get("start") != "2026-09-19T00:00:00Z" || query.Get("end") != "2026-09-22T00:00:00Z" {
				t.Errorf("body query = %v", query)
			}
		}
		fixtureName, ok := paths[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(fixture(t, fixtureName))
	}))
	defer server.Close()

	client, err := NewClient(ClientConfig{BaseURL: server.URL + "/developer", AccessToken: "contract-token", MaxRetries: 0})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	profile, err := client.Profile(ctx)
	if err != nil || profile.UserID != 1 {
		t.Fatalf("profile = %#v, err = %v", profile, err)
	}
	body, err := client.BodyMeasurementsPage(ctx, ListOptions{Limit: 1, Start: "2026-09-19T00:00:00Z", End: "2026-09-22T00:00:00Z"})
	if err != nil || len(body.Records) != 1 || body.Records[0].WeightKilogram != 70.4 {
		t.Fatalf("body = %#v, err = %v", body, err)
	}
	cycle, err := client.CyclesPage(ctx, ListOptions{Limit: 1})
	if err != nil || len(cycle.Records) != 1 || cycle.Records[0].Score == nil || cycle.Records[0].Score.Strain != 11.2 {
		t.Fatalf("cycle = %#v, err = %v", cycle, err)
	}
	recovery, err := client.RecoveryPage(ctx, ListOptions{Limit: 2})
	if err != nil || len(recovery.Records) != 2 || recovery.Records[0].Score == nil || recovery.Records[1].Score != nil || recovery.Records[1].ScoreState != "PENDING" {
		t.Fatalf("recovery = %#v, err = %v", recovery, err)
	}
	sleep, err := client.SleepPage(ctx, ListOptions{Limit: 1})
	if err != nil || len(sleep.Records) != 1 || sleep.Records[0].Score == nil || sleep.Records[0].Score.StageSummary == nil || sleep.Records[0].Score.StageSummary.TotalRemSleepTimeMilli != 5760000 {
		t.Fatalf("sleep = %#v, err = %v", sleep, err)
	}
	workouts, err := client.WorkoutsPage(ctx, ListOptions{Limit: 1})
	if err != nil || len(workouts.Records) != 1 || workouts.Records[0].Score == nil || workouts.Records[0].Score.Strain != 12.3 {
		t.Fatalf("workouts = %#v, err = %v", workouts, err)
	}
	for path := range paths {
		if seen[path] != 1 {
			t.Errorf("%s requested %d times, want once", path, seen[path])
		}
	}
}
