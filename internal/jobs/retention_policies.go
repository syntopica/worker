package jobs

import (
	"context"
	"maps"
	"time"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/store"
)

const (
	removedRetentionDays = 7
	removedUnackedHours  = 72
	removedParkedFloor   = 10 * time.Minute
)

func retentionPolicies(ctx context.Context, s *store.Store, cfg *config.Config) (map[string]config.QueuePolicy, error) {
	policies := make(map[string]config.QueuePolicy, len(cfg.Queues))
	maps.Copy(policies, cfg.Queues)
	names, err := queryStrings(ctx, s, "retention_queues", nil)
	for _, name := range names {
		if _, configured := policies[name]; !configured {
			policies[name] = config.QueuePolicy{
				RetentionDays: removedRetentionDays, UnackedTTLHours: removedUnackedHours, ParkedMinIdle: removedParkedFloor,
			}
		}
	}
	return policies, err
}
