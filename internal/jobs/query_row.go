package jobs

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/syntopica/worker/internal/store"
)

// queryRow reads one row into dest and reports whether there was one.
func queryRow(ctx context.Context, c store.Conn, name string, p params, dest ...any) (bool, error) {
	err := c.QueryRowContext(ctx, script(name), p.list()...).Scan(dest...)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("jobs: %s: %w", name, err)
	}

	return true, nil
}
