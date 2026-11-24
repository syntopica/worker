package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/store"
)

func settleUnsuccessful(
	ctx context.Context, tx store.Conn, cfg *config.Config, job *jobRow, report *CompletionReport, now float64,
) (string, error) {
	code := report.ErrorCode
	switch {
	case code == "rate_limited", code == "node_shutdown", code == "no_output" && job.Kind == kindInference && job.Pin == "":
		return stateQueued, requeueJob(ctx, tx, job.ID, code, now, now)
	case report.Outcome == outcomePreempted:
		return preemptJob(ctx, tx, cfg, job, now)
	case code == codeQuotaWall:
		return stateQueued, requeueAfterWall(ctx, tx, cfg, job, report, now)
	default:
		if code == "" {
			code = codeExecutorError
		}
		return failAttempt(ctx, tx, job, code, nil, now)
	}
}
