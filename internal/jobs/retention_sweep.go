package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/store"
)

type retentionSweep struct {
	queue         string
	now           float64
	parkedIdle    float64
	unackedBefore float64
	keepBefore    float64
}

func (sweep retentionSweep) apply(ctx context.Context, s *store.Store) (int, error) {
	steps := []struct {
		query  string
		apply  func(context.Context, store.Conn, string) error
		purges bool
	}{
		{"retention_splits", sweep.expireSplit, false},
		{"retention_unacked", sweep.expireSuccess, true},
		{"retention_payloads", deleteRetainedPayloads, true},
		{"retention_releases", releaseRetainedJob, false},
	}
	total := 0
	for _, step := range steps {
		for {
			count, err := sweep.batch(ctx, s, step.query, step.apply)
			if err != nil {
				return total, err
			}
			if step.purges {
				total += count
			}
			if count < retentionBatchLimit {
				break
			}
		}
	}
	return total, nil
}

func (sweep retentionSweep) batch(
	ctx context.Context, s *store.Store, script string, apply func(context.Context, store.Conn, string) error,
) (int, error) {
	return write(ctx, s, func(tx store.Conn) (int, error) {
		ids, err := queryStrings(ctx, tx, script, params{
			keyQueue: sweep.queue, keyLimit: retentionBatchLimit, "unacked_before": sweep.unackedBefore,
			"retained_before": sweep.keepBefore, "inspected_before": sweep.now - daySeconds,
		})
		if err != nil {
			return 0, err
		}
		for _, id := range ids {
			if err := apply(ctx, tx, id); err != nil {
				return 0, err
			}
		}
		return len(ids), nil
	})
}

func (sweep retentionSweep) expireSplit(ctx context.Context, tx store.Conn, id string) error {
	if err := run(ctx, tx, "ack_open_splits", params{keyID: id, keyNow: sweep.now}); err != nil {
		return err
	}
	return parkJob(ctx, tx, id, sweep.queue, sweep.parkedIdle, sweep.now)
}

func (sweep retentionSweep) expireSuccess(ctx context.Context, tx store.Conn, id string) error {
	job, err := loadJob(ctx, tx, id)
	if err != nil {
		return err
	}
	if err := run(ctx, tx, "expire_unacked", params{keyID: id, keyNow: sweep.now}); err != nil {
		return err
	}
	if err := addControlResult(ctx, tx, job, stateUnackedExpired, map[string]any{}, sweep.now); err != nil {
		return err
	}
	return deleteRetainedPayloads(ctx, tx, id)
}

func deleteRetainedPayloads(ctx context.Context, tx store.Conn, id string) error {
	return run(ctx, tx, "delete_payloads", params{keyID: id})
}

func releaseRetainedJob(ctx context.Context, tx store.Conn, id string) error {
	return run(ctx, tx, "release_job", params{keyID: id})
}
