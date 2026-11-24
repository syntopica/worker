package store

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func freshStore(t *testing.T, dir string) *Store {
	t.Helper()

	err := Migrate(t.Context(), dir)
	if err != nil {
		t.Fatal(err)
	}

	opened, err := Open(t.Context(), dir)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = opened.Close() })

	return opened
}

func scan[T any](t *testing.T, conn Conn, query string) T {
	t.Helper()

	var value T

	err := conn.QueryRowContext(t.Context(), query).Scan(&value)
	if err != nil {
		t.Fatalf("%s: %v", query, err)
	}

	return value
}

func tableNames(t *testing.T, conn Conn, schema string) map[string]bool {
	t.Helper()

	rows, err := conn.QueryContext(t.Context(), "SELECT name FROM "+schema+".sqlite_master WHERE type='table'")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = rows.Close() }()

	names := map[string]bool{}

	for rows.Next() {
		var name string

		err = rows.Scan(&name)
		if err != nil {
			t.Fatal(err)
		}

		names[name] = true
	}

	err = rows.Err()
	if err != nil {
		t.Fatal(err)
	}

	return names
}

func TestTwoFilesWithTheRequiredPragmas(t *testing.T) {
	t.Parallel()

	opened := freshStore(t, t.TempDir())

	if got := scan[string](t, opened, "PRAGMA journal_mode"); got != "wal" {
		t.Fatalf("journal_mode = %q", got)
	}

	checks := map[string]int{
		"PRAGMA synchronous":          2,
		"PRAGMA p.secure_delete":      1,
		"PRAGMA p.journal_size_limit": 0,
		"PRAGMA user_version":         Version,
	}
	for pragma, want := range checks {
		if got := scan[int](t, opened, pragma); got != want {
			t.Fatalf("%s = %d, want %d", pragma, got, want)
		}
	}

	payload := tableNames(t, opened, "p")
	if len(payload) != 2 || !payload["inputs"] || !payload["outputs"] {
		t.Fatalf("payload tables = %v", payload)
	}

	main := tableNames(t, opened, "main")
	if main["inputs"] || main["outputs"] || !main["jobs"] {
		t.Fatalf("main tables = %v", main)
	}
}

var errBoom = errors.New("boom")

func TestWriteRollsBackOnErrorAndCommitsOtherwise(t *testing.T) {
	t.Parallel()

	opened := freshStore(t, t.TempDir())
	insert := func(tx Conn) error {
		_, err := tx.ExecContext(t.Context(), "INSERT INTO nodes (name, report, updated) VALUES ('n', '{}', 0)")

		return err
	}

	err := opened.Write(t.Context(), func(tx Conn) error {
		return errors.Join(insert(tx), errBoom)
	})
	if !errors.Is(err, errBoom) {
		t.Fatalf("err = %v", err)
	}

	if got := scan[int](t, opened, "SELECT count(*) FROM nodes"); got != 0 {
		t.Fatalf("rows after rollback = %d", got)
	}

	err = opened.Write(t.Context(), insert)
	if err != nil {
		t.Fatal(err)
	}

	if got := scan[int](t, opened, "SELECT count(*) FROM nodes"); got != 1 {
		t.Fatalf("rows after commit = %d", got)
	}
}

func TestStateFilesArePrivate(t *testing.T) {
	t.Parallel()

	dir := filepath.Join(t.TempDir(), "state")
	opened := freshStore(t, dir)

	_, err := opened.ExecContext(t.Context(), "INSERT INTO nodes (name, report, updated) VALUES ('n', '{}', 0)")
	if err != nil {
		t.Fatal(err)
	}

	err = opened.Protect(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(dir)
	if err != nil || info.Mode().Perm() != privateDir {
		t.Fatalf("directory mode = %v, %v", info.Mode(), err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) < 4 {
		t.Fatalf("entries = %v, %v", entries, err)
	}

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil || info.Mode().Perm() != privateFile {
			t.Fatalf("%s mode = %v, %v", entry.Name(), info.Mode(), err)
		}
	}
}

func TestAnUnmigratedStoreIsRefused(t *testing.T) {
	t.Parallel()

	dir := filepath.Join(t.TempDir(), "new")

	_, err := Open(t.Context(), dir)

	var refused *NotMigratedError
	if !errors.As(err, &refused) || refused.Found != 0 || refused.Wanted != Version {
		t.Fatalf("err = %v", err)
	}

	if refused.Error() != "store schema version 0 is older than 10; run worker serve" {
		t.Fatalf("message = %q", refused.Error())
	}
}

func TestAReadCompletesWhileAnotherConnectionHoldsTheWriterLock(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writer := freshStore(t, dir)
	started := time.Now()

	err := writer.Write(t.Context(), func(Conn) error {
		reader, err := Open(t.Context(), dir)
		if err != nil {
			return err
		}

		defer func() { _ = reader.Close() }()

		var jobs int

		return reader.QueryRowContext(t.Context(), "SELECT count(*) FROM jobs").Scan(&jobs)
	})
	if err != nil {
		t.Fatal(err)
	}

	if time.Since(started) > 2*time.Second {
		t.Fatal("the read waited for the writer lock")
	}
}

func TestOpenFailsOnAnUnusableDirectory(t *testing.T) {
	t.Parallel()

	file := filepath.Join(t.TempDir(), "file")

	err := os.WriteFile(file, nil, privateFile)
	if err != nil {
		t.Fatal(err)
	}

	_, err = Open(t.Context(), filepath.Join(file, "state"))
	if err == nil {
		t.Fatal("a store under a regular file opened")
	}

	err = Migrate(t.Context(), filepath.Join(file, "state"))
	if err == nil {
		t.Fatal("a store under a regular file migrated")
	}
}

func TestACancelledContextRefusesTheTransaction(t *testing.T) {
	t.Parallel()

	opened := freshStore(t, t.TempDir())
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	err := opened.Write(ctx, func(Conn) error { return nil })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v", err)
	}

	var _ Conn = (*sql.Tx)(nil)
}
