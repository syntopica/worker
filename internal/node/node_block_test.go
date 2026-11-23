package node

import (
	"testing"
	"time"

	"github.com/syntopica/worker/internal/host"
)

func TestNodeBlockOrder(t *testing.T) {
	t.Parallel()

	settled := func() *memory {
		m := newMemory()
		m.normalSince = new(at(0))

		return m
	}
	unread := idleState()
	unread.Idle = nil

	cases := map[string]struct {
		change func(*memory)
		state  host.State
		down   bool
		want   reason
	}{
		"a failed drain first": {func(m *memory) { m.failedModel = testModel }, unread, true, drainBlocked},
		"unreadable host":      {nil, unread, true, hostStateUnreadable},
		"unknown pressure":     {nil, hostState(time.Hour, true, host.PressureUnknown), false, hostStateUnreadable},
		"pressure":             {nil, hostState(time.Hour, true, host.PressureWarn), false, memoryPressure},
		"typing is no block":   {nil, typingState(), false, ""},
		"recovering":           {func(m *memory) { m.normalSince = new(at(900)) }, idleState(), true, pressureRecovering},
		"never normal":         {func(m *memory) { m.normalSince = nil }, idleState(), false, pressureRecovering},
		"backend down":         {func(m *memory) { m.backoffUntil = at(2000) }, idleState(), true, backendDown},
		"pressure backoff":     {func(m *memory) { m.backoffUntil, m.transportUntil = at(2000), at(2000) }, idleState(), false, pressureBackoff},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			m := settled()
			if tc.change != nil {
				tc.change(m)
			}

			if got := nodeBlock(m, tc.state, at(1000), tc.down); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}
