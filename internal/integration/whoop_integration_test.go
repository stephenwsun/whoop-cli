//go:build integration

package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stephensun/whoop-cli/internal/runtime"
	"github.com/stephensun/whoop-cli/whoop"
)

func TestPublicAccountReads(t *testing.T) {
	if os.Getenv("WHOOP_IT") != "1" {
		t.Skip("set WHOOP_IT=1 to run live WHOOP integration tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	config, err := runtime.Load(ctx)
	if err != nil {
		t.Fatal(err)
	}
	client, err := config.API(ctx)
	if err != nil {
		t.Fatal(err)
	}
	profile, err := client.Profile(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if profile.UserID == "" {
		t.Fatal("WHOOP profile did not include user_id")
	}
	if _, err := client.Workouts(ctx, whoop.ListOptions{Limit: 1, Start: time.Now().UTC().Add(-24 * time.Hour).Format(time.RFC3339)}); err != nil {
		t.Fatal(err)
	}
}
