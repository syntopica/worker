package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/store"
)

// ReconcileLostPayloads fails non-terminal jobs whose inputs are missing after restore.
func ReconcileLostPayloads(ctx context.Context, s *store.Store, now float64) (int, error) {
	count, err := write(ctx, s, func(tx store.Conn) (int, error) {
		ids, err := queryStrings(ctx, tx, "lost_payloads", nil)
		if err != nil {
			return 0, err
		}
		for _, id := range ids {
			if err := failPayloadLost(ctx, tx, id, now); err != nil {
				return 0, err
			}
		}
		return len(ids), nil
	})
	if err != nil {
		return 0, err
	}
	return count, nil
}
