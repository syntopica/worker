package host_test

import (
	"testing"
	"time"

	"github.com/syntopica/worker/internal/host"
)

func TestHoldPowerSourceReusesARecentReading(t *testing.T) {
	t.Parallel()

	sample, now, advance := scripted(
		host.State{OnAC: new(true), Pressure: host.PressureNormal},
		host.State{Pressure: host.PressureNormal},
		host.State{Pressure: host.PressureNormal},
	)
	held := host.HoldPowerSource(sample, now)

	held(t.Context())
	advance(2 * time.Minute)

	if state := held(t.Context()); state.OnAC == nil || !*state.OnAC {
		t.Errorf("on_ac = %v, want the two-minute-old reading", state.OnAC)
	}

	advance(time.Second)

	if state := held(t.Context()); state.OnAC != nil {
		t.Error("a reading older than two minutes was reused")
	}
}

func TestHoldPowerSourceKeepsABatteryReading(t *testing.T) {
	t.Parallel()

	sample, now, _ := scripted(host.State{OnAC: new(false)}, host.State{})
	held := host.HoldPowerSource(sample, now)

	held(t.Context())

	if state := held(t.Context()); state.OnAC == nil || *state.OnAC {
		t.Errorf("on_ac = %v, want battery held", state.OnAC)
	}
}
