package store

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// Backup takes an online snapshot of the metadata database only. The
// destination is a new 0600 file, published only after the copy succeeds.
func Backup(ctx context.Context, dir, dest string) error {
	source, err := filepath.Abs(filepath.Join(dir, metaFile))
	if err != nil {
		return fmt.Errorf("resolve metadata path: %w", err)
	}
	info, err := os.Stat(source)
	if errors.Is(err, fs.ErrNotExist) || (err == nil && !info.Mode().IsRegular()) {
		return ErrBackupSourceMissing
	}
	if err != nil {
		return fmt.Errorf("inspect metadata: %w", err)
	}
	dest, err = filepath.Abs(dest)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrBackupDestination, err)
	}
	if err := checkBackupDestination(filepath.Dir(source), dest); err != nil {
		return err
	}
	if err := Migrate(ctx, dir); err != nil {
		return err
	}
	return writeBackup(ctx, source, dest)
}

func checkBackupDestination(state, dest string) error {
	if _, err := os.Lstat(dest); !errors.Is(err, fs.ErrNotExist) {
		if err == nil {
			err = fs.ErrExist
		}
		return fmt.Errorf("%w: %w", ErrBackupDestination, err)
	}
	// SQLite can create sidecars after this command starts. Do not publish a
	// backup at one of those paths, including through an alias of state.
	switch filepath.Base(dest) {
	case metaFile + "-wal", metaFile + "-shm", metaFile + "-journal",
		payloadFile + "-wal", payloadFile + "-shm", payloadFile + "-journal":
	default:
		return nil
	}
	stateInfo, stateErr := os.Stat(state)
	destDir, destErr := os.Stat(filepath.Dir(dest))
	if err := errors.Join(stateErr, destErr); err != nil {
		return fmt.Errorf("%w: %w", ErrBackupDestination, err)
	}
	if os.SameFile(stateInfo, destDir) {
		return ErrBackupDestination
	}
	return nil
}

func writeBackup(ctx context.Context, source, dest string) (err error) {
	temp, err := os.CreateTemp(filepath.Dir(dest), ".worker-backup-*")
	if err != nil {
		return fmt.Errorf("%w: %w", ErrBackupDestination, err)
	}
	defer func() {
		if closeErr := temp.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close scratch backup: %w", closeErr))
		}
		if cleanup := os.Remove(temp.Name()); cleanup != nil && !errors.Is(cleanup, fs.ErrNotExist) {
			err = errors.Join(err, fmt.Errorf("remove scratch backup: %w", cleanup))
		}
	}()
	if err := copyMetadata(ctx, source, temp.Name()); err != nil {
		return err
	}
	return publishBackup(ctx, temp, dest, os.Link)
}
