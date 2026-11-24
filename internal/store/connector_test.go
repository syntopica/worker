package store

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestMigrationUsesStatePathWithURICharacters(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	dir := filepath.Join(root, "state?tag=with space#100%25&extra")

	if err := Migrate(t.Context(), dir); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(root, "state")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("migration created a file at the truncated state path: %v", err)
	}

	for _, name := range []string{metaFile, payloadFile} {
		info, err := os.Stat(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		if info.Size() == 0 {
			t.Errorf("migration left %s empty", name)
		}
	}

	opened, err := Open(t.Context(), dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = opened.Close() })

	if got := scan[int](t, opened, "PRAGMA busy_timeout"); got != 10000 {
		t.Errorf("busy_timeout = %d, want 10000", got)
	}
}

func TestMigrationUsesLiteralFilePrefixInRelativeStatePath(t *testing.T) {
	t.Chdir(t.TempDir())

	dir := "file:state?label=example"
	if err := Migrate(t.Context(), dir); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(filepath.Join(dir, payloadFile))
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() == 0 {
		t.Fatal("migration left the payload file empty")
	}
}
