package tasks

import (
	"context"
	"fmt"
	"sync"
)

// RunSlots runs the task loop in several slots, each with its own
// coordinator link, until the context ends. A stopped slot kills its runner
// and hands its job back before it returns.
func RunSlots(ctx context.Context, slots []*Loop) error {
	var running sync.WaitGroup

	for _, slot := range slots {
		running.Go(func() { _ = slot.Run(ctx) })
	}

	running.Wait()

	return fmt.Errorf("task slots stopped: %w", ctx.Err())
}
