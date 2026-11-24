package jobs

import (
	"context"
	"database/sql"
	"errors"

	"github.com/syntopica/worker/internal/store"
)

func ownedJob(ctx context.Context, tx store.Conn, producer *string, id string) (*jobRow, error) {
	job, err := loadJob(ctx, tx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, notFound()
	}
	if err != nil {
		return nil, err
	}
	if producer != nil && job.Producer != *producer {
		return nil, notFound()
	}
	return job, nil
}
