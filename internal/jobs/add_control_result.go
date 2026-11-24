package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/store"
)

// addControlResult tells the producer something about a job that is not an
// output; the detail carries allowlisted fields only (spec 9).
func addControlResult(
	ctx context.Context, c store.Conn, job *jobRow, control string, detail map[string]any, now float64,
) error {
	encoded, err := encodeValue(detail)
	if err != nil {
		return err
	}

	return run(ctx, c, "add_control_result", params{
		keyResultID: newID(), keyID: job.ID, keyProducer: job.Producer, keyQueue: job.Queue, "control": control,
		"detail": encoded, keyNow: now,
	})
}
