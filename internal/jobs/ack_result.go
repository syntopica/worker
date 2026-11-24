package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/contract"
	"github.com/syntopica/worker/internal/store"
)

// AckResult acknowledges a producer's result and may update its rating or decline a split.
func AckResult(
	ctx context.Context, s *store.Store, cfg *config.Config, producer, id string, ack contract.Ack, now float64,
) error {
	if ack.Rating != "" && !KnownRating(ack.Rating) {
		return bad("unknown_rating")
	}
	return s.Write(ctx, func(tx store.Conn) error {
		var control *string
		found, err := queryRow(ctx, tx, "ack_control",
			params{keyResultID: ack.ResultID, keyID: id, keyProducer: producer}, &control)
		if err != nil {
			return err
		}
		if !found {
			return notFound()
		}
		job, err := ownedJob(ctx, tx, &producer, id)
		if err != nil {
			return err
		}
		return applyAck(ctx, tx, cfg, job, ack, control, now)
	})
}
