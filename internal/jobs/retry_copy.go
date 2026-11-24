package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/store"
)

func retryCopy(
	ctx context.Context, tx store.Conn, cfg *config.Config, job *jobRow, sub *submission, digest, model string, now float64,
) (string, error) {
	if err := run(ctx, tx, "ack_all_results", params{keyID: job.ID, keyNow: now}); err != nil {
		return "", err
	}
	if err := markJobAcked(ctx, tx, job, now); err != nil {
		return "", err
	}
	if err := checkOutstanding(ctx, tx, cfg, job.Producer, sub.Queue); err != nil {
		return "", err
	}
	id := newID()
	if err := insertJob(ctx, tx, id, job.Producer, sub, digest, model, job.ID, now); err != nil {
		return "", err
	}
	return id, failEmptyLadder(ctx, tx, cfg, id, now)
}
