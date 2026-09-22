package auth

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestFileStoreCreatesOwnerOnlyCredentialFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission bits are not applied on Windows")
	}
	dir := filepath.Join(t.TempDir(), "whoop")
	store := FileStore{Path: filepath.Join(dir, "credentials.json")}
	ctx := context.Background()
	for _, key := range []string{"refresh_token", "client_id"} {
		if err := store.Set(ctx, key, "value-"+key); err != nil {
			t.Fatalf("set %s: %v", key, err)
		}
		info, err := os.Stat(store.Path)
		if err != nil {
			t.Fatal(err)
		}
		if got := info.Mode().Perm(); got != 0o600 {
			t.Fatalf("after writing %s, credential file mode = %o, want 600", key, got)
		}
	}
	dirInfo, err := os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got := dirInfo.Mode().Perm(); got != 0o700 {
		t.Fatalf("credential directory mode = %o, want 700", got)
	}
	leftovers, err := filepath.Glob(filepath.Join(dir, ".credentials-*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(leftovers) != 0 {
		t.Fatalf("temporary credential files left behind: %v", leftovers)
	}
}
