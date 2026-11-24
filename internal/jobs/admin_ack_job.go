package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/store"
)

// AdminAckJob acknowledges a job's open control results on its producer's behalf.
func AdminAckJob(ctx context.Context, s *store.Store, cfg *config.Config, id string, now float64) error {
	return s.Write(ctx, func(tx store.Conn) error {
		job, err := ownedJob(ctx, tx, nil, id)
		if err != nil {
			return err
		}
		return ackOnlyControls(ctx, tx, cfg, job, now)
	})
}
