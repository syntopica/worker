package tasks

import (
	"context"
	"sync"
	"sync/atomic"
)

// RunOnDemand drains an on-demand profile's jobs in several slots and
// returns the total run; each slot stops on its own once its lease comes
// back empty.
func RunOnDemand(ctx context.Context, slots []*OnDemand) int {
	var (
		running sync.WaitGroup
		total   atomic.Int64
	)

	for _, slot := range slots {
		running.Go(func() { total.Add(int64(slot.Run(ctx))) })
	}

	running.Wait()

	return int(total.Load())
}
