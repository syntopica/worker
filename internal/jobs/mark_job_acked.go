package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/store"
)

func markJobAcked(ctx context.Context, tx store.Conn, job *jobRow, now float64) error {
	if err := run(ctx, tx, "mark_job_acked", params{keyID: job.ID, keyNow: now}); err != nil {
		return err
	}
	return deleteSensitivePayloads(ctx, tx, job)
}
