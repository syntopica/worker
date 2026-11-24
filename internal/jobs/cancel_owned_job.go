package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/store"
)

func cancelOwnedJob(ctx context.Context, tx store.Conn, job *jobRow, now float64) (string, error) {
	if !liveState(job.State) && job.State != stateQueued && job.State != stateSplitRequested {
		return job.State, nil
	}
	if err := run(ctx, tx, "cancel_job", params{keyID: job.ID, keyNow: now}); err != nil {
		return "", err
	}
	return stateCancelled, deleteSensitivePayloads(ctx, tx, job)
}
