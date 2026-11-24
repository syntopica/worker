package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
)

// connect creates the state directory and both database files owner-only,
// then returns a single-connection pool over them; it never touches the
// schema. SQLite gives later -wal and -shm files the mode of their database.
func connect(dir string) (*sql.DB, error) {
	if err := os.MkdirAll(dir, privateDir); err != nil {
		return nil, fmt.Errorf("create state directory: %w", err)
	}
	for _, name := range []string{metaFile, payloadFile} {
		//nolint:gosec // the state directory is the operator's own configuration
		file, err := os.OpenFile(filepath.Join(dir, name), os.O_CREATE|os.O_RDWR, privateFile)
		if err != nil {
			return nil, fmt.Errorf("create %s: %w", name, err)
		}
		if err := file.Close(); err != nil {
			return nil, fmt.Errorf("create %s: %w", name, err)
		}
	}
	db := sql.OpenDB(connector{dir: dir})
	db.SetMaxOpenConns(1)
	return db, nil
}
