package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/store"
)

func preemptJob(ctx context.Context, tx store.Conn, cfg *config.Config, job *jobRow, now float64) (string, error) {
	if job.Parked {
		runs := job.ParkedRuns + 1
		if runs >= cfg.MaxParkedRuns {
			return stateFailed, finishFailed(ctx, tx, job, "preemption_exhausted", stateFailed, nil, now)
		}
		return stateQueued, run(ctx, tx, "preempt_parked", params{keyID: job.ID, keyNow: now, "runs": runs})
	}
	count, state := job.Preemptions+1, stateQueued
	if count >= cfg.SplitAfterPreemptions {
		state = stateSplitRequested
	}
	err := run(ctx, tx, "preempt_job", params{keyID: job.ID, keyState: state, keyNow: now, "preemptions": count})
	if err != nil || state != stateSplitRequested {
		return state, err
	}
	return state, addControlResult(ctx, tx, job, state, map[string]any{"preemptions": count}, now)
}
