package cli_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/syntopica/worker/internal/store"
)

func TestBackupCommandExitStatuses(t *testing.T) {
	root := newInstance(t, "127.0.0.1:0")
	dest := filepath.Join(t.TempDir(), "backup.sqlite3")
	assertBackupExit(t, dest, 2, "worker: no metadata store\n")

	state := filepath.Join(root, "state")
	if err := store.Migrate(t.Context(), state); err != nil {
		t.Fatal(err)
	}
	assertBackupExit(t, dest, 0, "")
	assertBackupExit(t, dest, 2, "worker: backup destination already exists\n")

	if err := os.WriteFile(filepath.Join(state, "meta.sqlite3"), []byte("unparseable fixture"), 0o600); err != nil {
		t.Fatal(err)
	}
	assertBackupExit(t, filepath.Join(t.TempDir(), "another.sqlite3"), 1, "worker: backup failed\n")
}

func assertBackupExit(t *testing.T, dest string, wantCode int, wantDiagnostic string) {
	t.Helper()
	code, stdout, stderr := invoke(t, "", "backup", dest)
	if code != wantCode || stdout != "" || stderr != wantDiagnostic {
		t.Fatalf("backup = %d, %q, %q", code, stdout, stderr)
	}
}
