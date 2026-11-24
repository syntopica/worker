package store

import (
	"context"
	"errors"
	"fmt"
)

// Migrate creates or upgrades the store in a state directory, for the entry
// points that own it; an open never migrates.
func Migrate(ctx context.Context, dir string) error {
	db, err := connect(dir)
	if err != nil {
		return err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Join(fmt.Errorf("begin migration: %w", err), db.Close())
	}
	if err := migrateSchema(ctx, tx); err != nil {
		return errors.Join(err, tx.Rollback(), db.Close())
	}
	if err := tx.Commit(); err != nil {
		return errors.Join(fmt.Errorf("commit migration: %w", err), db.Close())
	}
	if err := db.Close(); err != nil {
		return fmt.Errorf("close migrated store: %w", err)
	}
	return protectFiles(dir)
}
