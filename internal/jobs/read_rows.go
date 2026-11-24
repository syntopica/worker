package jobs

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/syntopica/worker/internal/store"
)

func readRows[T any](
	ctx context.Context, c store.Conn, query string, p params, scan func(*sql.Rows) (T, error),
) (out []T, err error) {
	rows, err := c.QueryContext(ctx, query, p.list()...)
	if err != nil {
		return nil, fmt.Errorf("jobs: query: %w", err)
	}
	defer func() { err = errors.Join(err, rows.Close()) }()

	out = []T{}
	for rows.Next() {
		item, scanErr := scan(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		out = append(out, item)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("jobs: query: %w", err)
	}

	return out, nil
}
