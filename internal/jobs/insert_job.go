package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/store"
)

// insertJob stores one admitted job and its input inside the caller's
// transaction; the input goes to the payload file.
func insertJob(
	ctx context.Context, c store.Conn, id, producer string, sub *submission, digest, model, retryOf string, now float64,
) error {
	body, err := encodeValue(map[string]any{keyInput: sub.Input})
	if err != nil {
		return err
	}
	var exclude any
	if len(sub.Exclude) > 0 {
		exclude, err = encodeValue(sub.Exclude)
		if err != nil {
			return err
		}
	}

	return run(ctx, c, "insert_job", params{
		keyID: id, keyProducer: producer, keyQueue: sub.Queue, keyKind: sub.Kind, keyKey: sub.Key, "digest": digest,
		keyPriority: sub.Priority, keyPrivacy: sub.Privacy, "model": model, keyMaxAttempts: sub.MaxAttempts,
		keyNow: now, "deadline": sub.Deadline, "parent_id": sub.ParentID, keyTier: sub.Tier,
		"retry_of": nullIfEmpty(retryOf), "exclude": exclude, keyBody: body,
	})
}
