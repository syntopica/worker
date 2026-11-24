package jobs

import (
	"context"
	"database/sql"

	"github.com/syntopica/worker/internal/store"
)

// queryAll reads every row before it returns, so the caller may query again
// on the same connection. The result is never nil.
func queryAll[T any](
	ctx context.Context, c store.Conn, name string, p params, scan func(*sql.Rows) (T, error),
) ([]T, error) {
	return readRows(ctx, c, script(name), p, scan)
}
