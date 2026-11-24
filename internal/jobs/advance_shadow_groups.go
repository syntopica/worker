package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/store"
)

const shadowBatchLimit = 50

// AdvanceShadowGroups queues judges for finished groups and records their verdicts.
func AdvanceShadowGroups(ctx context.Context, s *store.Store, cfg *config.Config, now float64) (int, error) {
	err := s.Write(ctx, func(tx store.Conn) error {
		return run(ctx, tx, "cancel_stale_shadows", params{keyNow: now, keySince: now - daySeconds})
	})
	if err != nil {
		return 0, err
	}
	steps := []func(store.Conn, int) (int, error){
		func(tx store.Conn, limit int) (int, error) { return submitJudgesBatch(ctx, tx, cfg, now, limit) },
		func(tx store.Conn, limit int) (int, error) { return recordJudgementsBatch(ctx, tx, now, limit) },
	}
	moved := 0
	for _, step := range steps {
		for {
			done, err := write(ctx, s, func(tx store.Conn) (int, error) { return step(tx, shadowBatchLimit) })
			if err != nil {
				return moved, err
			}
			moved += done
			if done < shadowBatchLimit {
				break
			}
		}
	}
	return moved, nil
}
