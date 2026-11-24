package jobs

import (
	"context"
	"maps"
	"slices"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/policy"
	"github.com/syntopica/worker/internal/store"
)

// OpenRouterHeadroom returns today's use, the strictest key cap and capped queues.
func OpenRouterHeadroom(ctx context.Context, s *store.Store, cfg *config.Config, now float64) (map[string]any, error) {
	spent, err := openrouterSpentToday(ctx, s, now)
	if err != nil {
		return nil, err
	}
	total := 0
	for _, count := range spent {
		total += count
	}
	var keyCap *int
	for name := range cfg.Queues {
		queue := cfg.Queues[name]
		if queue.OpenRouter == nil || queue.OpenRouter.DailyKeyCap == nil {
			continue
		}
		current := queue.OpenRouter.DailyKeyCap
		if keyCap == nil || *current < *keyCap {
			keyCap = current
		}
	}
	capped := slices.Sorted(maps.Keys(policy.CappedQueues(cfg, spent)))
	if capped == nil {
		capped = []string{}
	}
	return map[string]any{"spent_today": total, "key_cap": keyCap, "capped": capped}, nil
}
