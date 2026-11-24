package jobs

import (
	"context"
	"database/sql"

	"github.com/syntopica/worker/internal/store"
)

// queryStrings reads a one-column result.
func queryStrings(ctx context.Context, c store.Conn, name string, p params) ([]string, error) {
	return queryAll(ctx, c, name, p, func(rows *sql.Rows) (string, error) {
		var value string
		err := scanInto(rows, &value)

		return value, err
	})
}
