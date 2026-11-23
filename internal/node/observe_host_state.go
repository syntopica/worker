package node

import (
	"time"

	"github.com/syntopica/worker/internal/host"
)

// observeHostState folds one host sample into the node's memory: since when
// pressure is normal, and for how many samples it has been pressed.
func observeHostState(m *memory, state host.State, now time.Time) {
	switch {
	case state.Pressure != host.PressureNormal:
		m.normalSince = nil
	case m.normalSince == nil:
		m.normalSince = &now
	}

	if state.Pressure.Pressed() {
		m.pressureStreak++

		return
	}

	m.pressureStreak = 0
	m.unloaded = false
}
