package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/store"
)

// HeartbeatAttempt renews a lease owned by the attempt and changes its live state.
func HeartbeatAttempt(
	ctx context.Context, s *store.Store, attempt string, generation int64, node string, draining bool, now float64,
) error {
	state := stateRunning
	if draining {
		state = stateDraining
	}
	return s.Write(ctx, func(tx store.Conn) error {
		id, err := checkFence(ctx, tx, attempt, generation, node)
		if err != nil {
			return err
		}
		return run(ctx, tx, "heartbeat", params{keyID: id, keyState: state, keyExpires: now + leaseTTL, keyNow: now})
	})
}
