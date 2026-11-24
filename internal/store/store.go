package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"runtime"
)

// Store is an already migrated store: one connection over the metadata file
// with the payload file attached as p. Only the metadata file is ever backed
// up (spec 9), which is why content lives in a file of its own.
type Store struct {
	Conn

	db    *sql.DB
	dir   string
	marks *backupMarks
}

// Open never migrates and never begins a transaction; it only reads
// user_version and refuses an older store with a NotMigratedError: the entry
// points run Migrate first.
func Open(ctx context.Context, dir string) (*Store, error) {
	db, err := connect(dir)
	if err != nil {
		return nil, err
	}
	var found int
	if err := db.QueryRowContext(ctx, "PRAGMA user_version").Scan(&found); err != nil {
		return nil, errors.Join(fmt.Errorf("read store version: %w", err), db.Close())
	}
	if found < Version {
		return nil, errors.Join(&NotMigratedError{Found: found, Wanted: Version}, db.Close())
	}
	opened := &Store{Conn: db, db: db, dir: dir, marks: newBackupMarks()}
	if err := opened.Protect(ctx); err != nil {
		return nil, errors.Join(err, db.Close())
	}
	return opened, nil
}

// Protect makes the state files owner-only and, on macOS, keeps the payload
// file out of Time Machine. Sidecar files appear after the first write, so a
// long-lived store calls it again from its periodic pass.
func (s *Store) Protect(ctx context.Context) error {
	if err := protectFiles(s.dir); err != nil {
		return err
	}
	if runtime.GOOS == "darwin" {
		s.marks.exclude(ctx, s.dir)
	}
	return nil
}

// Write runs one short write transaction that takes the writer lock up
// front: it commits when work returns nil and rolls back otherwise.
func (s *Store) Write(ctx context.Context, work func(tx Conn) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	if err := work(tx); err != nil {
		return errors.Join(err, tx.Rollback())
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// Close releases the connection.
func (s *Store) Close() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("close store: %w", err)
	}
	return nil
}
