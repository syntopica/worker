package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/contract"
	"github.com/syntopica/worker/internal/store"
)

func applyAck(
	ctx context.Context, tx store.Conn, cfg *config.Config, job *jobRow, ack contract.Ack, control *string, now float64,
) error {
	if err := run(ctx, tx, "ack_result", params{keyResultID: ack.ResultID, keyNow: now, "rating": nullIfEmpty(ack.Rating)}); err != nil {
		return err
	}
	if control == nil || *control == stateFailed || *control == stateExpired || *control == stateUnackedExpired {
		return markJobAcked(ctx, tx, job, now)
	}
	if *control == stateSplitRequested && ack.Decline && job.State == stateSplitRequested {
		return parkJob(ctx, tx, job.ID, job.Queue, cfg.Queues[job.Queue].ParkedMinIdle.Seconds(), now)
	}
	return nil
}
