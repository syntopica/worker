package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/store"
)

// CancelJob cancels unfinished work and fences its lease; a nil producer denotes an administrator.
func CancelJob(ctx context.Context, s *store.Store, producer *string, id string, now float64) (string, error) {
	state, err := write(ctx, s, func(tx store.Conn) (string, error) {
		job, err := ownedJob(ctx, tx, producer, id)
		if err != nil {
			return "", err
		}
		return cancelOwnedJob(ctx, tx, job, now)
	})
	if err != nil {
		return "", err
	}
	return state, nil
}
