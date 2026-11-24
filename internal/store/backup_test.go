package store

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBackupIncludesCommittedWALWithoutPayloads(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	seedUncheckpointedBackup(t, dir)

	dest := filepath.Join(t.TempDir(), "backup ?#%.sqlite3")
	if err := Backup(t.Context(), dir, dest); err != nil {
		t.Fatal(err)
	}
	assertPrivateBackupFile(t, dest)

	snapshot, err := sql.Open("sqlite", backupURI(dest, "mode=ro"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = snapshot.Close() })
	if got := scan[string](t, snapshot, `SELECT name FROM nodes`); got != "fixture-node" {
		t.Fatalf("backup node = %q", got)
	}
	if tables := tableNames(t, snapshot, "main"); tables["inputs"] || tables["outputs"] {
		t.Fatalf("backup includes payload tables: %v", tables)
	}
}

func seedUncheckpointedBackup(t *testing.T, dir string) {
	t.Helper()
	opened := freshStore(t, dir)
	for _, query := range []string{
		`PRAGMA wal_autocheckpoint=0`,
		`PRAGMA wal_checkpoint(TRUNCATE)`,
		`INSERT INTO nodes VALUES ('fixture-node', '{}', 1)`,
		`INSERT INTO p.inputs VALUES ('fixture-job', 'fixture-content-must-stay-in-payloads')`,
	} {
		if _, err := opened.ExecContext(t.Context(), query); err != nil {
			t.Fatal(err)
		}
	}
	wal, err := os.Stat(filepath.Join(dir, metaFile+"-wal"))
	if err != nil || wal.Size() == 0 {
		t.Fatalf("committed WAL fixture: %v, %v", wal, err)
	}
}

func assertPrivateBackupFile(t *testing.T, dest string) {
	t.Helper()
	info, err := os.Stat(dest)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("backup mode = %v, error = %v", info, err)
	}
	contents, err := os.ReadFile(dest)
	if err != nil || bytes.Contains(contents, []byte("fixture-content-must-stay-in-payloads")) {
		t.Fatal("backup missing or contains a payload")
	}
}

func TestFailedBackupLeavesNoDestinationOrScratchFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	source, dest := filepath.Join(dir, "missing.sqlite3"), filepath.Join(dir, "backup.sqlite3")
	if err := writeBackup(t.Context(), source, dest); err == nil {
		t.Fatal("backup read a missing source database")
	}
	if entries, err := os.ReadDir(dir); err != nil || len(entries) != 0 {
		t.Fatalf("failed backup left files: %v, %v", entries, err)
	}
}

func TestBackupRetriesLocksUntilCanceled(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	source, dest := filepath.Join(dir, "source.sqlite3"), filepath.Join(t.TempDir(), "backup.sqlite3")
	database, err := sql.Open("sqlite", source)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	conn, err := database.Conn(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if _, err = conn.ExecContext(t.Context(), `CREATE TABLE fixture (value TEXT); BEGIN EXCLUSIVE`); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 150*time.Millisecond)
	defer cancel()
	if err = writeBackup(ctx, source, dest); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("backup stopped before waiting for the lock: %v", err)
	}
	if entries, readErr := os.ReadDir(filepath.Dir(dest)); readErr != nil || len(entries) != 0 {
		t.Fatalf("canceled backup left files: %v, %v", entries, readErr)
	}
}

func TestBackupRequiresExistingStore(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	dir, dest := filepath.Join(root, "state"), filepath.Join(root, "backup.sqlite3")
	if err := Backup(t.Context(), dir, dest); !errors.Is(err, ErrBackupSourceMissing) {
		t.Fatalf("backup with missing store: %v", err)
	}
	if entries, err := os.ReadDir(root); err != nil || len(entries) != 0 {
		t.Fatalf("backup with missing store created files: %v, %v", entries, err)
	}
}

func TestBackupPreservesAnExistingDestination(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	freshStore(t, dir)
	dest := filepath.Join(t.TempDir(), "backup.sqlite3")
	if err := os.WriteFile(dest, []byte("keep me"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := Backup(t.Context(), dir, dest); !errors.Is(err, fs.ErrExist) {
		t.Fatalf("backup with existing destination: %v", err)
	}
	raw, err := os.ReadFile(dest)
	if err != nil || string(raw) != "keep me" {
		t.Fatalf("existing destination changed: %q, %v", raw, err)
	}
}

func TestBackupRefusesDestinationsReservedForSQLite(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	freshStore(t, dir)
	alias := filepath.Join(t.TempDir(), "state-alias")
	if err := os.Symlink(dir, alias); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(alias, payloadFile+"-journal")
	if err := Backup(t.Context(), dir, dest); !errors.Is(err, ErrBackupDestination) {
		t.Fatalf("backup into a reserved file: %v", err)
	}
	if _, err := os.Stat(dest); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("reserved file was created: %v", err)
	}
}
