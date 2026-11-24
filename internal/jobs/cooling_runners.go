package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/store"
)

// coolingRunners merges quota cooldowns with concurrency caps. Both are
// checked inside the same write transaction as the new lease.
func coolingRunners(ctx context.Context, tx store.Conn, cfg *config.Config, now float64) (map[string]bool, error) {
	runners, err := queryStrings(ctx, tx, "cooling_runners", params{keyNow: now})
	if err != nil {
		return nil, err
	}
	cooling := make(map[string]bool, len(runners))
	for _, runner := range runners {
		cooling[runner] = true
	}
	if len(cfg.RunnerMaxConcurrent) == 0 {
		return cooling, nil
	}
	live, err := queryStrings(ctx, tx, "live_profiles", nil)
	if err != nil {
		return nil, err
	}
	count := map[string]int{}
	for _, name := range live {
		if profile, ok := cfg.Profiles[name]; ok {
			count[profile.Runner]++
		}
	}
	for runner, limit := range cfg.RunnerMaxConcurrent {
		if count[runner] >= limit {
			cooling[runner] = true
		}
	}
	return cooling, nil
}
