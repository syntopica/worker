package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"time"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

const (
	backupStepPages = 128
	backupLockWait  = 250 * time.Millisecond
	sqliteCodeMask  = 0xff
)

var errNoSQLiteBackup = errors.New("online SQLite backup unavailable")

func copyMetadata(ctx context.Context, source, dest string) (err error) {
	// Open metadata on its own connection. That connection cannot see the
	// adjacent payload database, even by an attached schema name.
	database, err := sql.Open("sqlite", backupURI(source, "mode=ro"))
	if err != nil {
		return fmt.Errorf("open metadata: %w", err)
	}
	defer func() { err = errors.Join(err, database.Close()) }()
	conn, err := database.Conn(ctx)
	if err != nil {
		return fmt.Errorf("connect to metadata: %w", err)
	}
	defer func() { err = errors.Join(err, conn.Close()) }()
	if err := conn.Raw(func(driver any) error {
		return copyMetadataPages(ctx, driver, backupURI(dest, "mode=rw"))
	}); err != nil {
		return fmt.Errorf("copy metadata: %w", err)
	}
	return nil
}

func copyMetadataPages(ctx context.Context, driver any, dest string) (err error) {
	online, ok := driver.(interface {
		NewBackup(destURI string) (*sqlite.Backup, error)
	})
	if !ok {
		return errNoSQLiteBackup
	}
	backup, err := online.NewBackup(dest)
	if err != nil {
		return fmt.Errorf("start SQLite backup: %w", err)
	}
	defer func() { err = errors.Join(err, backup.Finish()) }()
	for ctx.Err() == nil {
		more, err := backup.Step(backupStepPages)
		if backupLocked(err) {
			select {
			case <-ctx.Done():
			case <-time.After(backupLockWait):
			}
			continue
		}
		if err != nil {
			return fmt.Errorf("copy SQLite pages: %w", err)
		}
		if !more {
			return nil
		}
	}
	return fmt.Errorf("backup interrupted: %w", ctx.Err())
}

func backupLocked(err error) bool {
	failure, ok := errors.AsType[*sqlite.Error](err)
	if !ok {
		return false
	}
	code := failure.Code() & sqliteCodeMask
	return code == sqlite3.SQLITE_BUSY || code == sqlite3.SQLITE_LOCKED
}

func backupURI(path, query string) string {
	uri := url.URL{Scheme: "file", Path: filepath.ToSlash(path), RawQuery: query}
	return uri.String()
}
