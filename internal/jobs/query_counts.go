package jobs

import (
	"context"
	"database/sql"

	"github.com/syntopica/worker/internal/store"
)

// queryCounts reads a (name, count) result.
func queryCounts(ctx context.Context, c store.Conn, name string, p params) (map[string]int, error) {
	type pair struct {
		name  string
		count int
	}
	pairs, err := queryAll(ctx, c, name, p, func(rows *sql.Rows) (pair, error) {
		var found pair
		scanErr := scanInto(rows, &found.name, &found.count)

		return found, scanErr
	})
	counts := make(map[string]int, len(pairs))
	for _, found := range pairs {
		counts[found.name] = found.count
	}

	return counts, err
}
