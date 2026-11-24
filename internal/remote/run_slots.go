package remote

import (
	"context"
	"fmt"
	"sync"
)

// RunSlots runs the loop in several slots, each with its own coordinator
// link, until the context ends. A stopped slot hands its running job back
// before it returns.
func RunSlots(ctx context.Context, slots []*Loop) error {
	var running sync.WaitGroup

	for _, slot := range slots {
		running.Go(func() { _ = slot.Run(ctx) })
	}

	running.Wait()

	return fmt.Errorf("remote slots stopped: %w", ctx.Err())
}
