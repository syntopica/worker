package jobs

import (
	"context"
	"slices"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/policy"
	"github.com/syntopica/worker/internal/store"
)

// failEmptyLadder fails a just-admitted job whose excluded providers leave
// its ladder empty, as no_eligible_executor. Only an exclusion that removes
// every provider the ladder had counts: a ladder already empty without it
// waits as before. No other executor is tried.
func failEmptyLadder(ctx context.Context, c store.Conn, cfg *config.Config, id string, now float64) error {
	job, err := loadJob(ctx, c, id)
	if err != nil {
		return err
	}
	ladder := policy.LadderProviders(cfg, job.Queue, job.Privacy, job.Model)
	excluded, err := excludedProviders(job.Exclude)
	if err != nil {
		return err
	}
	if len(excluded) == 0 || len(ladder) == 0 {
		return nil
	}
	for provider := range ladder {
		if !slices.Contains(excluded, provider) {
			return nil
		}
	}

	return finishFailed(ctx, c, job, "no_eligible_executor", stateFailed, nil, now)
}
