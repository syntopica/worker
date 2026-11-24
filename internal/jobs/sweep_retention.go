package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/store"
)

// SweepRetention applies the retention table in bounded transactions and counts deleted payloads.
func SweepRetention(ctx context.Context, s *store.Store, cfg *config.Config, now float64) (int, error) {
	policies, err := retentionPolicies(ctx, s, cfg)
	if err != nil {
		return 0, err
	}
	total := 0
	for queue := range policies {
		policy := policies[queue]
		sweep := retentionSweep{
			queue: queue, now: now, parkedIdle: policy.ParkedMinIdle.Seconds(),
			unackedBefore: now - float64(policy.UnackedTTLHours)*hourSeconds,
			keepBefore:    now - float64(policy.RetentionDays)*daySeconds,
		}
		count, err := sweep.apply(ctx, s)
		total += count
		if err != nil {
			return total, err
		}
	}
	if total > 0 {
		return total, run(ctx, s, "checkpoint_payloads", nil)
	}
	return total, nil
}
