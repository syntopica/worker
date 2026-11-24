package node

import "time"

const (
	maxBackoff               = 2 * time.Hour
	firstTransportBackoff    = 30 * time.Second
	maxTransportBackoff      = 30 * time.Minute
	maxTransportBackoffShift = 6 // Six doublings of 30 seconds reach the 30-minute cap.
)

// settleAttempt folds an attempt's outcome into the node's memory and
// reports whether the node should rest before asking for more work.
//
// A pressure release starts a backoff of 15 minutes, doubling per consecutive
// pressure release up to 2 hours; a success resets it. A model becomes the
// node's own only when the attempt ran on it: a failed attempt may never
// have loaded it. Consecutive transport failures hold the node back 30 s,
// doubling up to 30 minutes, so a backend that lists models but cannot chat
// does not charge an attempt every rest.
func settleAttempt(m *memory, model string, wasResident bool, how outcome, why reason, now time.Time) bool {
	if !wasResident && how.ran() {
		m.owned[model] = true
	}

	if how == drainFailed {
		m.failedModel = model
	}

	if how == succeeded {
		m.backoff = firstBackoff
		m.backoffUntil = time.Time{}
	}

	transport := how == failed && why == transportError
	if transport {
		m.transportFailures++
		shift := min(m.transportFailures-1, maxTransportBackoffShift)
		m.transportUntil = now.Add(min(firstTransportBackoff<<shift, maxTransportBackoff))
	} else {
		m.transportFailures = 0
	}

	if why == memoryPressure && (how == preempted || how == drainFailed) {
		m.owned[model] = true
		m.backoffUntil = now.Add(m.backoff)
		m.backoff = min(m.backoff*2, maxBackoff)
		m.pressureStreak = max(m.pressureStreak, pressureSamples)
	}

	return transport
}
