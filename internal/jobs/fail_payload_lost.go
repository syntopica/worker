package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/store"
)

func failPayloadLost(ctx context.Context, tx store.Conn, id string, now float64) error {
	job, err := loadJob(ctx, tx, id)
	if err != nil {
		return err
	}
	if err := finishFailed(ctx, tx, job, "payload_lost", stateFailed, nil, now); err != nil {
		return err
	}
	return run(ctx, tx, "delete_payloads", params{keyID: id})
}
