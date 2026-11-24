package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/store"
)

// ReadCosts returns grouped execution metadata since the given time.
func ReadCosts(ctx context.Context, s *store.Store, since float64) ([]map[string]any, error) {
	return queryMaps(ctx, s, "read_costs", params{keySince: since})
}
