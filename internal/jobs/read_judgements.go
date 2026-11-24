package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/store"
)

// ReadJudgements returns grouped execution metadata since the given time.
func ReadJudgements(ctx context.Context, s *store.Store, since float64) ([]map[string]any, error) {
	return queryMaps(ctx, s, "read_judgements", params{keySince: since})
}
