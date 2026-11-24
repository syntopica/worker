package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/store"
)

// ReadResultRatings returns grouped execution metadata since the given time.
func ReadResultRatings(ctx context.Context, s *store.Store, since float64) ([]map[string]any, error) {
	return queryMaps(ctx, s, "read_result_ratings", params{keySince: since})
}
