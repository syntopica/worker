package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/store"
)

const (
	maxActivityHours = 168
	hourlyUpTo       = 48
	coarseBucket     = 6 * hourSeconds
)

// ReadActivity returns aggregate attempt counts, wall time and tokens in time buckets.
func ReadActivity(ctx context.Context, s *store.Store, now, hours float64) (map[string]any, error) {
	hours = min(max(hours, 1), maxActivityHours)
	since := now - hours*hourSeconds
	bucket := hourSeconds
	if hours > hourlyUpTo {
		bucket = coarseBucket
	}
	rows, err := queryMaps(ctx, s, "read_activity", params{"bucket": bucket, keySince: since, keyNow: now})
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		row["sampling"] = row["sampling"] == int64(1)
	}
	return map[string]any{keySince: since, "bucket_s": bucket, "rows": rows}, nil
}
