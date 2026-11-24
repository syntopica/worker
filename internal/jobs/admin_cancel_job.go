package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/store"
)

// AdminCancelJob cancels a job and records the administrator's action in the same transaction.
func AdminCancelJob(ctx context.Context, s *store.Store, id, principal string, now float64) (string, error) {
	return write(ctx, s, func(tx store.Conn) (string, error) {
		job, err := ownedJob(ctx, tx, nil, id)
		if err != nil {
			return "", err
		}
		state, err := cancelOwnedJob(ctx, tx, job, now)
		if err != nil {
			return "", err
		}
		return state, recordAudit(ctx, tx, id, job.Privacy, principal, "cancel", now)
	})
}
