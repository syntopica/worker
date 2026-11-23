package node

import (
	"time"

	"github.com/syntopica/worker/internal/host"
)

// idleTime is the sampled idle time, zero when it could not be read.
func idleTime(state host.State) time.Duration {
	if state.Idle == nil {
		return 0
	}

	return *state.Idle
}
