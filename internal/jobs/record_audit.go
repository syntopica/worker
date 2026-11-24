package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/store"
)

func recordAudit(ctx context.Context, tx store.Conn, id, privacy, principal, action string, now float64) error {
	return run(ctx, tx, "record_audit", params{
		"action": action, keyID: id, keyPrivacy: privacy, "principal": principal, keyNow: now,
	})
}
