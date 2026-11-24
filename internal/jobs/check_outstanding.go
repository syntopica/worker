package jobs

import (
	"context"
	"net/http"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/fault"
	"github.com/syntopica/worker/internal/store"
)

// checkOutstanding refuses with outstanding_limit when the producer already
// holds the queue's limit of unacknowledged jobs.
func checkOutstanding(ctx context.Context, c store.Conn, cfg *config.Config, producer, queue string) error {
	var held int
	if _, err := queryRow(ctx, c, "outstanding", params{keyProducer: producer, keyQueue: queue}, &held); err != nil {
		return err
	}
	if held >= cfg.Queues[queue].MaxOutstanding {
		return fault.New(http.StatusTooManyRequests, "outstanding_limit")
	}

	return nil
}
