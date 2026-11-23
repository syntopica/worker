package node

import (
	"time"

	"github.com/syntopica/worker/internal/host"
)

// pressureRecovery is how long pressure must stay normal before new work.
const pressureRecovery = 2 * time.Minute

// nodeBlock is why the node asks for no work, or "".
func nodeBlock(m *memory, state host.State, now time.Time, backendIsDown bool) reason {
	if m.failedModel != "" {
		return drainBlocked
	}

	if why := releaseReason(state, "active_ok"); why != "" {
		return why
	}

	switch {
	case m.normalSince == nil || now.Sub(*m.normalSince) < pressureRecovery:
		return pressureRecovering
	case backendIsDown:
		return backendDown
	case now.Before(m.backoffUntil):
		return pressureBackoff
	case now.Before(m.transportUntil):
		return transportBackoff
	default:
		return ""
	}
}
