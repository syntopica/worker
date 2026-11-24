package tasks

import (
	"context"

	"github.com/syntopica/worker/internal/link"
)

// OnDemand drains one on-demand profile's jobs through one slot. There is
// no host check: the owner started this run, so neither activity nor idle
// time holds it back.
type OnDemand struct {
	// Loop is the slot; its Sample and Usage are not used.
	Loop *Loop
	// Profile is the on-demand profile to run.
	Profile string
	// OverridePins also takes jobs pinned to another profile.
	OverridePins bool
	// CreatedBefore takes only jobs created before this Unix time, when set.
	CreatedBefore *float64
}

// Run runs jobs until none is left for the profile, its runner is resting
// or the context ends, and returns how many it ran. Held completions are
// retried once more at the end, since no later lease will carry them.
func (o *OnDemand) Run(ctx context.Context) int {
	ran := 0

	for ctx.Err() == nil {
		lease := o.Loop.Link.LeaseTask(ctx, &link.TaskRequest{
			Profile:       o.Profile,
			OverridePins:  o.OverridePins,
			CreatedBefore: o.CreatedBefore,
		})
		if lease == nil {
			break
		}

		o.Loop.runLeased(ctx, o.Loop.Config(), lease)

		ran++
	}

	o.Loop.Link.Flush(context.WithoutCancel(ctx))

	return ran
}
