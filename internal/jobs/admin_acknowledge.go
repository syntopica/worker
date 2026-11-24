package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/store"
)

// AdminAcknowledge acknowledges control results and records the action in the same transaction.
func AdminAcknowledge(ctx context.Context, s *store.Store, cfg *config.Config, id, principal string, now float64) (string, error) {
	return write(ctx, s, func(tx store.Conn) (string, error) {
		job, err := ownedJob(ctx, tx, nil, id)
		if err != nil {
			return "", err
		}
		if err := ackOnlyControls(ctx, tx, cfg, job, now); err != nil {
			return "", err
		}
		if err := recordAudit(ctx, tx, id, job.Privacy, principal, "ack", now); err != nil {
			return "", err
		}
		return JobState(ctx, tx, id)
	})
}
