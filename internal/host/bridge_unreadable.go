package host

import (
	"context"
	"time"
)

const bridgeWindow = 30 * time.Second

// BridgeUnreadable wraps a sampler so a failed idle or pressure read reuses
// one at most thirty seconds old.
//
// Under load the host readers occasionally fail one sample; on 2026-10-01
// that caused 17 of 20 preemptions. Each reader runs in turn with a five
// second timeout, so a sample in which they all time out lasts about 25 s:
// the window covers one such sample, and a reading that old delays a release
// by at most that long. The idle time is reused as read, never advanced, so it
// cannot cross the return threshold on its own.
func BridgeUnreadable(sample Sampler, now func() time.Time) Sampler {
	var (
		idle       *time.Duration
		idleAt     time.Time
		pressure   = PressureUnknown
		pressureAt time.Time
	)

	return func(ctx context.Context) State {
		state := sample(ctx)
		at := now()

		switch {
		case state.Idle != nil:
			idle, idleAt = state.Idle, at
		case idle != nil && at.Sub(idleAt) <= bridgeWindow:
			state.Idle = idle
		}

		switch {
		case state.Pressure != PressureUnknown:
			pressure, pressureAt = state.Pressure, at
		case pressure != PressureUnknown && at.Sub(pressureAt) <= bridgeWindow:
			state.Pressure = pressure
		}

		return state
	}
}
