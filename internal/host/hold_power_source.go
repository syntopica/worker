package host

import (
	"context"
	"time"
)

const powerHold = 2 * time.Minute

// HoldPowerSource wraps a sampler so a failed power read reuses a reading at
// most two minutes old.
//
// "pmset -g ps" failed nine times on 2026-09-30 even after one immediate
// retry, and each failure preempted the running attempt as
// host_state_unreadable. The power source changes rarely; trusting a
// two-minute-old reading costs at most two minutes on battery.
func HoldPowerSource(sample Sampler, now func() time.Time) Sampler {
	var (
		last   *bool
		lastAt time.Time
	)

	return func(ctx context.Context) State {
		state := sample(ctx)
		at := now()

		switch {
		case state.OnAC != nil:
			last, lastAt = state.OnAC, at
		case last != nil && at.Sub(lastAt) <= powerHold:
			state.OnAC = last
		}

		return state
	}
}
