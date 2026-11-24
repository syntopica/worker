package store

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestBackupPublicationFallsBackWithoutOverwriting(t *testing.T) {
	t.Parallel()
	snapshot := backupPublicationFixture(t)
	dest := filepath.Join(t.TempDir(), "backup.sqlite3")
	if err := publishBackup(t.Context(), snapshot, dest, unsupportedBackupLink); err != nil {
		t.Fatal(err)
	}
	if err := publishBackup(t.Context(), snapshot, dest, unsupportedBackupLink); !errors.Is(err, fs.ErrExist) {
		t.Fatalf("expected existing destination refusal, got %v", err)
	}
	info, err := os.Stat(dest)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("backup mode = %v, error = %v", info, err)
	}
	contents, err := os.ReadFile(dest)
	if err != nil || string(contents) != "complete snapshot" {
		t.Fatalf("published backup contents = %q, error = %v", contents, err)
	}
}

func TestBackupFallbackCleansFailedCopy(t *testing.T) {
	t.Parallel()
	snapshot := backupPublicationFixture(t)
	if err := snapshot.Close(); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(t.TempDir(), "backup.sqlite3")
	if err := publishBackup(t.Context(), snapshot, dest, unsupportedBackupLink); err == nil {
		t.Fatal("backup from a closed snapshot returned success")
	}
	if _, err := os.Lstat(dest); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("failed backup left a destination: %v", err)
	}
}

func backupPublicationFixture(t *testing.T) *os.File {
	t.Helper()
	path := filepath.Join(t.TempDir(), "snapshot.sqlite3")
	if err := os.WriteFile(path, []byte("complete snapshot"), 0o600); err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = file.Close() })
	return file
}

func unsupportedBackupLink(oldname, newname string) error {
	return &os.LinkError{Op: "link", Old: oldname, New: newname, Err: syscall.EOPNOTSUPP}
}
