package store

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"syscall"
)

func publishBackup(ctx context.Context, snapshot *os.File, dest string, link func(oldname, newname string) error) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("publish backup: %w", err)
	}
	err := link(snapshot.Name(), dest)
	if err == nil {
		return nil
	}
	if !hardLinkUnavailable(err) {
		return fmt.Errorf("%w: %w", ErrBackupDestination, err)
	}
	return publishBackupCopy(ctx, snapshot, dest)
}

func hardLinkUnavailable(err error) bool {
	return errors.Is(err, errors.ErrUnsupported) || errors.Is(err, syscall.EPERM) ||
		errors.Is(err, syscall.ENOTSUP) || errors.Is(err, syscall.EOPNOTSUPP)
}

func publishBackupCopy(ctx context.Context, snapshot *os.File, dest string) error {
	// The operator supplies the destination. Exclusive creation refuses an
	// existing file or symlink, even if another writer creates it first.
	target, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_EXCL, privateFile)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrBackupDestination, err)
	}
	_, copyErr := io.Copy(target, snapshot)
	if copyErr == nil {
		copyErr = target.Sync()
	}
	if err := errors.Join(copyErr, target.Close(), ctx.Err()); err != nil {
		failure := fmt.Errorf("write backup: %w", err)
		if cleanup := os.Remove(dest); cleanup != nil && !errors.Is(cleanup, fs.ErrNotExist) {
			return errors.Join(failure, fmt.Errorf("remove incomplete backup: %w", cleanup))
		}
		return failure
	}
	return nil
}
