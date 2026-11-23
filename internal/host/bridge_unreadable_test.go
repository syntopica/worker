package host_test

import (
	"context"
	"testing"
	"time"

	"github.com/syntopica/worker/internal/host"
)

// scripted returns a sampler that yields each state in turn and a clock the
// test advances.
func scripted(states ...host.State) (sample host.Sampler, now func() time.Time, advance func(time.Duration)) {
	at := time.Unix(1000, 0)
	next := 0

	sample = func(context.Context) host.State {
		state := states[next]
		next++

		return state
	}

	return sample, func() time.Time { return at }, func(d time.Duration) { at = at.Add(d) }
}

func TestBridgeUnreadableReusesARecentReading(t *testing.T) {
	t.Parallel()

	read := host.State{Idle: new(40 * time.Second), OnAC: new(true), Pressure: host.PressureNormal}
	sample, now, advance := scripted(read, host.State{Pressure: host.PressureUnknown}, host.State{Pressure: host.PressureUnknown})
	bridged := host.BridgeUnreadable(sample, now)

	bridged(t.Context())
	advance(30 * time.Second)

	state := bridged(t.Context())
	if state.Idle == nil || *state.Idle != 40*time.Second {
		t.Errorf("idle = %v, want the 40s reading reused as read", state.Idle)
	}

	if state.Pressure != host.PressureNormal {
		t.Errorf("pressure = %s, want normal reused", state.Pressure)
	}

	if state.OnAC != nil {
		t.Error("the power source is not bridged")
	}

	advance(time.Second)

	state = bridged(t.Context())
	if state.Idle != nil || state.Pressure != host.PressureUnknown {
		t.Errorf("a reading older than 30s was reused: %+v", state)
	}
}

func TestBridgeUnreadableHasNothingToReuseAtFirst(t *testing.T) {
	t.Parallel()

	sample, now, _ := scripted(host.State{Pressure: host.PressureUnknown})

	state := host.BridgeUnreadable(sample, now)(t.Context())
	if state.Idle != nil || state.Pressure != host.PressureUnknown {
		t.Errorf("state = %+v, want unreadable", state)
	}
}
