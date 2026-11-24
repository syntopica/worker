package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/store"
)

func deleteSensitivePayloads(ctx context.Context, tx store.Conn, job *jobRow) error {
	if !Sensitive(job.Privacy) {
		return nil
	}
	return run(ctx, tx, "delete_payloads", params{keyID: job.ID})
}
