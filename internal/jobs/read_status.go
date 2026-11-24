package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/store"
)

// ReadStatus returns queue counts and recent totals, node status, cooldowns and failures.
func ReadStatus(ctx context.Context, s *store.Store, now float64) (map[string]any, error) {
	queues, err := statusQueues(ctx, s, now)
	if err != nil {
		return nil, err
	}
	nodes, err := statusNodes(ctx, s, now)
	if err != nil {
		return nil, err
	}
	cooldowns, err := statusCooldowns(ctx, s, now)
	if err != nil {
		return nil, err
	}
	failures, err := queryMaps(ctx, s, "status_failures", nil)
	if err != nil {
		return nil, err
	}
	return map[string]any{"queues": queues, "nodes": nodes, "cooldowns": cooldowns, "recent_failures": failures}, nil
}

func statusCooldowns(ctx context.Context, s *store.Store, now float64) (map[string]any, error) {
	rows, err := queryMaps(ctx, s, "status_cooldowns", params{keyNow: now})
	if err != nil {
		return nil, err
	}
	cooldowns := make(map[string]any, len(rows))
	for _, row := range rows {
		cooldowns[text(row[keyRunner])] = row["remaining"]
	}
	return cooldowns, nil
}
