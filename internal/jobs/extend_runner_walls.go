package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/store"
)

// ExtendRunnerWalls records active quota resets without shortening any existing cooldown.
func ExtendRunnerWalls(ctx context.Context, s *store.Store, walls map[string]float64, now float64) error {
	return s.Write(ctx, func(tx store.Conn) error {
		for runner, reset := range walls {
			if reset <= now {
				continue
			}
			if err := run(ctx, tx, "extend_runner_wall", params{keyRunner: runner, "until": reset}); err != nil {
				return err
			}
		}
		return nil
	})
}
