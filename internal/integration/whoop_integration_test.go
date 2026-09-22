//go:build integration

package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stephenwsun/whoop-cli/internal/runtime"
	"github.com/stephenwsun/whoop-cli/whoop"
)

func TestPublicAccountReads(t *testing.T) {
	if os.Getenv("WHOOP_IT") != "1" {
		t.Skip("set WHOOP_IT=1 to run live WHOOP integration tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
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
	if profile.UserID <= 0 {
		t.Fatal("WHOOP profile did not include user_id")
	}
	start := time.Now().UTC().Add(-24 * time.Hour).Format(time.RFC3339)
	opts := whoop.ListOptions{Limit: 5, Start: start}
	if _, err := client.BodyMeasurements(ctx, opts); err != nil {
		t.Fatal("body measurements: ", err)
	}
	if _, err := client.Cycles(ctx, opts); err != nil {
		t.Fatal("cycles: ", err)
	}
	if _, err := client.Recovery(ctx, opts); err != nil {
		t.Fatal("recovery: ", err)
	}
	if _, err := client.Sleep(ctx, opts); err != nil {
		t.Fatal("sleep: ", err)
	}
	if _, err := client.Workouts(ctx, opts); err != nil {
		t.Fatal("workouts: ", err)
	}
}
