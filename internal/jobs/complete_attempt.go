package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/store"
)

// CompleteAttempt settles a fenced attempt and returns its job's new state.
func CompleteAttempt(
	ctx context.Context, s *store.Store, cfg *config.Config, attempt string, generation int64, node string,
	report *CompletionReport, now float64,
) (string, error) {
	state, err := write(ctx, s, func(tx store.Conn) (string, error) {
		id, err := checkFence(ctx, tx, attempt, generation, node)
		if err != nil {
			return "", err
		}
		job, err := loadJob(ctx, tx, id)
		if err != nil {
			return "", err
		}
		settlement := completedReport(cfg, job, report)
		if err := endAttempt(ctx, tx, attempt, &settlement, now); err != nil {
			return "", err
		}
		if settlement.Outcome != stateSucceeded {
			return settleUnsuccessful(ctx, tx, cfg, job, &settlement, now)
		}
		return acceptOutput(ctx, tx, cfg, job, &settlement, now)
	})
	if err != nil {
		return "", err
	}
	return state, nil
}
