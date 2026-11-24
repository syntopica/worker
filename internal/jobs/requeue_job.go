package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/store"
)

func requeueJob(ctx context.Context, tx store.Conn, id, code string, now, due float64) error {
	return run(ctx, tx, "requeue_job", params{keyID: id, keyCode: code, keyNow: now, "due": due})
}
