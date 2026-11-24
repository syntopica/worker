package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/store"
)

// ExpireLeases reclaims lost leases and expires queued jobs past deadline.
func ExpireLeases(ctx context.Context, s *store.Store, now float64) (int, error) {
	count, err := write(ctx, s, func(tx store.Conn) (int, error) {
		lost, err := queryStrings(ctx, tx, "expired_leases", params{keyNow: now})
		if err != nil {
			return 0, err
		}
		for _, id := range lost {
			if err := loseAttempt(ctx, tx, id, now); err != nil {
				return 0, err
			}
		}
		return len(lost), expireDeadlines(ctx, tx, now)
	})
	if err != nil {
		return 0, err
	}
	return count, nil
}

func loseAttempt(ctx context.Context, tx store.Conn, id string, now float64) error {
	job, err := loadJob(ctx, tx, id)
	if err != nil {
		return err
	}
	if err := run(ctx, tx, "lose_attempt", params{keyID: id, keyAttempt: job.LeaseAttempt, keyNow: now}); err != nil {
		return err
	}
	if job.Attempts+1 >= job.MaxAttempts {
		return finishFailed(ctx, tx, job, "lease_lost", stateFailed, nil, now)
	}
	return run(ctx, tx, "requeue_lost", params{keyID: id, keyNow: now})
}

func expireDeadlines(ctx context.Context, tx store.Conn, now float64) error {
	ids, err := queryStrings(ctx, tx, "expired_deadlines", params{keyNow: now})
	if err != nil {
		return err
	}
	for _, id := range ids {
		job, err := loadJob(ctx, tx, id)
		if err != nil {
			return err
		}
		if err := finishFailed(ctx, tx, job, "deadline", stateExpired, nil, now); err != nil {
			return err
		}
	}
	return nil
}
