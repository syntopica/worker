package jobs

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/syntopica/worker/internal/store"
)

func queryMapRows(ctx context.Context, c store.Conn, query string, p params) ([]map[string]any, error) {
	return readRows(ctx, c, query, p, func(rows *sql.Rows) (map[string]any, error) {
		names, err := rows.Columns()
		if err != nil {
			return nil, fmt.Errorf("jobs: query columns: %w", err)
		}
		values := make([]any, len(names))
		targets := make([]any, len(names))
		for i := range values {
			targets[i] = &values[i]
		}
		err = scanInto(rows, targets...)
		row := make(map[string]any, len(names))
		for i, column := range names {
			row[column] = values[i]
			if raw, ok := values[i].([]byte); ok {
				row[column] = string(raw)
			}
		}

		return row, err
	})
}
