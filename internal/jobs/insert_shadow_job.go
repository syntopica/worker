package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/store"
)

func insertShadowJob(
	ctx context.Context, tx store.Conn, job *jobRow, input map[string]any, member *shadowMember, now float64,
) (string, error) {
	body, err := encodeValue(map[string]any{keyInput: input})
	if err != nil {
		return "", errInvalidInput
	}
	id := newID()
	err = run(ctx, tx, "insert_shadow_job", params{
		keyID: id, keyQueue: job.Queue, keyPrivacy: job.Privacy, keyModel: member.model, keyNow: now,
		keyKey: member.pin + ":" + member.group, keyTier: job.Tier, keyGroup: member.group, "pin": member.pin, keyBody: body,
		keyKind: member.kind, keyProducer: member.producer, keyMaxAttempts: member.maxAttempts,
	})
	return id, err
}
