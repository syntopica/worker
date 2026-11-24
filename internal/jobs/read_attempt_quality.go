package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/store"
)

// ReadAttemptQuality returns grouped execution metadata since the given time.
func ReadAttemptQuality(ctx context.Context, s *store.Store, since float64) ([]map[string]any, error) {
	return queryMaps(ctx, s, "read_attempt_quality", params{keySince: since})
}
