package jobs

import (
	"database/sql"
	"fmt"
)

// scanInto scans the current row and wraps the failure once.
func scanInto(rows *sql.Rows, dest ...any) error {
	if err := rows.Scan(dest...); err != nil {
		return fmt.Errorf("jobs: scan: %w", err)
	}

	return nil
}
