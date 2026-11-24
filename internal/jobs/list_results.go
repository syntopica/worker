package jobs

import (
	"context"
	"database/sql"

	"github.com/syntopica/worker/internal/store"
)

const maxResultLimit = 100

// ListResults returns the producer's unacknowledged results in sequence order.
func ListResults(
	ctx context.Context, s *store.Store, producer, queue string, after int64, limit int,
) ([]JobResult, error) {
	return queryAll(ctx, s, "list_results", params{
		keyProducer: producer, keyQueue: queue, "after": after, keyLimit: max(1, min(limit, maxResultLimit)),
	}, scanResult)
}

func scanResult(rows *sql.Rows) (JobResult, error) {
	var result JobResult
	err := scanInto(rows, result.dest()...)
	return result, err
}
