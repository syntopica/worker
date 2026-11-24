package jobs

import (
	"context"
	"math"

	"github.com/syntopica/worker/internal/store"
)

const (
	firstRetryDelay  = 30
	maxRetryDelay    = 900
	unavailableDelay = 600
)

// failAttempt charges a job one attempt and applies retry backoff.
func failAttempt(
	ctx context.Context, tx store.Conn, job *jobRow, code string, detail map[string]any, now float64,
) (string, error) {
	if err := run(ctx, tx, "charge_attempt", params{keyID: job.ID}); err != nil {
		return "", err
	}
	if job.Attempts+1 >= job.MaxAttempts {
		return stateFailed, finishFailed(ctx, tx, job, code, stateFailed, detail, now)
	}
	delay := min(math.Ldexp(firstRetryDelay, job.Attempts), maxRetryDelay)
	if code == "runner_unavailable" {
		delay = max(delay, unavailableDelay)
	}
	return stateQueued, requeueJob(ctx, tx, job.ID, code, now, now+delay)
}
