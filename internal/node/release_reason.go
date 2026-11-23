package node

import (
	"time"

	"github.com/syntopica/worker/internal/host"
)

// userReturn is the idle time below which the user has returned.
const userReturn = 10 * time.Second

// releaseReason is why a running attempt must be released, or "" (spec 7).
// An unreadable signal is a reason: failure counts as busy.
func releaseReason(state host.State, runWhen string) reason {
	switch {
	case state.Idle == nil || state.OnAC == nil || state.Pressure == host.PressureUnknown:
		return hostStateUnreadable
	case !*state.OnAC:
		return onBattery
	case state.Pressure != host.PressureNormal:
		return memoryPressure
	case runWhen == runWhenIdle && *state.Idle < userReturn:
		return userActive
	default:
		return ""
	}
}
