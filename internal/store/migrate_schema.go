package store

import (
	"context"
	_ "embed"
	"fmt"
	"strings"
)

//go:embed sql/schema.sql
var schemaSQL string

// addedColumnsSQL adds, in the order they were introduced, the columns that
// came after the first schema; the last one is schema version 9.
//
//go:embed sql/added_columns.sql
var addedColumnsSQL string

//go:embed sql/indexes.sql
var indexesSQL string

//go:embed sql/has_column.sql
var hasColumnSQL string

//go:embed sql/mark_deleted_payloads.sql
var markDeletedPayloadsSQL string

// migrateSchema is forward and idempotent: it adds what is missing and never
// drops or rewrites data. A store from before payloads_deleted gains the
// column and has the finished jobs whose input is already gone marked, so the
// first sweep after the upgrade does not revisit them. user_version is set
// last, in the same transaction.
func migrateSchema(ctx context.Context, tx Conn) error {
	err := runStatements(ctx, tx, schemaSQL)
	if err != nil {
		return err
	}
	for statement := range strings.SplitSeq(addedColumnsSQL, ";") {
		err = addColumn(ctx, tx, strings.TrimSpace(statement))
		if err != nil {
			return err
		}
	}
	err = runStatements(ctx, tx, indexesSQL)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, fmt.Sprintf("PRAGMA user_version=%d", Version))
	if err != nil {
		return fmt.Errorf("set store version: %w", err)
	}
	return nil
}

func runStatements(ctx context.Context, tx Conn, script string) error {
	for statement := range strings.SplitSeq(script, ";") {
		if strings.TrimSpace(statement) == "" {
			continue
		}
		_, err := tx.ExecContext(ctx, statement)
		if err != nil {
			return fmt.Errorf("migrate schema: %w", err)
		}
	}
	return nil
}

// addColumn runs one "ALTER TABLE t ADD COLUMN c ..." unless c is there.
func addColumn(ctx context.Context, tx Conn, statement string) error {
	words := strings.Fields(statement)
	if len(words) <= columnWord {
		return nil
	}
	table, column := words[tableWord], words[columnWord]
	var present int
	err := tx.QueryRowContext(ctx, hasColumnSQL, table, column).Scan(&present)
	if err != nil {
		return fmt.Errorf("inspect %s: %w", table, err)
	}
	if present > 0 {
		return nil
	}
	_, err = tx.ExecContext(ctx, statement)
	if err != nil {
		return fmt.Errorf("add %s.%s: %w", table, column, err)
	}
	if column != "payloads_deleted" {
		return nil
	}
	_, err = tx.ExecContext(ctx, markDeletedPayloadsSQL)
	if err != nil {
		return fmt.Errorf("mark deleted payloads: %w", err)
	}
	return nil
}

// Positions of the table and the column in an ALTER TABLE statement.
const (
	tableWord  = 2
	columnWord = 5
)
