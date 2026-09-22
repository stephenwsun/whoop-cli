package version

import "testing"

func TestCurrentAndStringExposeBuildMetadata(t *testing.T) {
	oldVersion, oldCommit, oldDate := Version, Commit, Date
	t.Cleanup(func() { Version, Commit, Date = oldVersion, oldCommit, oldDate })
	Version, Commit, Date = "v1.2.3", "abc123", "2026-09-22T00:00:00Z"

	info := Current()
	if info.Version != Version || info.Commit != Commit || info.Date != Date {
		t.Fatalf("Current() = %#v", info)
	}
	if got := info.String(); got != "v1.2.3 (commit abc123, built 2026-09-22T00:00:00Z)" {
		t.Fatalf("String() = %q", got)
	}
	if got := UserAgent(); got != "whoop-cli/v1.2.3" {
		t.Fatalf("UserAgent() = %q", got)
	}
}
