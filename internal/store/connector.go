package store

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"

	"modernc.org/sqlite"
)

// connector opens one connection to the metadata file with the payload file
// attached as p, so every connection of the pool sees both.
type connector struct {
	dir string
}

var errNoExec = errors.New("sqlite connection cannot execute statements")

// Connect waits at most ten seconds for the writer lock and takes it at
// BEGIN, as the Python store did.
func (c connector) Connect(ctx context.Context) (driver.Conn, error) {
	dir, err := filepath.Abs(c.dir)
	if err != nil {
		return nil, fmt.Errorf("resolve state directory: %w", err)
	}
	c.dir = dir
	path := filepath.ToSlash(filepath.Join(dir, metaFile))
	if path[0] != '/' {
		path = "/" + path
	}
	dsn := url.URL{
		Scheme:   "file",
		Path:     path,
		RawQuery: "_txlock=immediate&_busy_timeout=10000",
	}
	conn, err := c.Driver().Open(dsn.String())
	if err != nil {
		return nil, fmt.Errorf("open metadata: %w", err)
	}
	if err := c.prepare(ctx, conn); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return conn, nil
}

// Driver is a private driver: nothing registered elsewhere reaches it.
func (connector) Driver() driver.Driver {
	return &sqlite.Driver{}
}

func (c connector) prepare(ctx context.Context, conn driver.Conn) error {
	execer, ok := conn.(driver.ExecerContext)
	if !ok {
		return errNoExec
	}
	payloads := []driver.NamedValue{{Ordinal: 1, Value: filepath.Join(c.dir, payloadFile)}}
	if _, err := execer.ExecContext(ctx, "ATTACH DATABASE ? AS p", payloads); err != nil {
		return fmt.Errorf("attach payloads: %w", err)
	}
	for _, pragma := range []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA synchronous=FULL",
		"PRAGMA p.journal_mode=WAL",
		"PRAGMA p.synchronous=FULL",
		"PRAGMA p.secure_delete=ON",
		"PRAGMA p.journal_size_limit=0",
	} {
		if _, err := execer.ExecContext(ctx, pragma, nil); err != nil {
			return fmt.Errorf("%s: %w", pragma, err)
		}
	}
	return nil
}
