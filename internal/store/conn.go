package store

import (
	"context"
	"database/sql"
)

// Conn is what a query needs: the store itself outside a transaction, or
// the transaction a write runs in.
type Conn interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}
