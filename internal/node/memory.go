package node

import "time"

// memory is what the node remembers between iterations.
type memory struct {
	// normalSince is when pressure last became normal; nil while it is not.
	normalSince *time.Time
	// pressureStreak counts consecutive pressed samples.
	pressureStreak int
	// unloaded says the pressure episode already had its unload.
	unloaded bool
	// failedModel is the model whose drain failed, "" when none.
	failedModel string
	// owned are the models this node loaded: the only ones it unloads.
	owned map[string]bool
	// batteryTried and activeTried hold the models already sent an unload
	// in the current battery episode or active spell.
	batteryTried map[string]bool
	activeTried  map[string]bool
	// backoffUntil ends the backoff after a pressure release; backoff is
	// the length of the next one.
	backoffUntil time.Time
	backoff      time.Duration
	// transportFailures counts consecutive transport failures, which hold
	// the node back until transportUntil.
	transportFailures int
	transportUntil    time.Time
	// drainRetryAt is when a failed drain's unload may be sent again.
	drainRetryAt time.Time
	// lastResident is the previous residency sample, nil before the first.
	lastResident map[string]bool
	// residency holds the loads and unloads of the last hour.
	residency []residencyEvent
}

// newMemory is the memory of a node that has just started.
func newMemory() *memory {
	return &memory{
		owned:        map[string]bool{},
		batteryTried: map[string]bool{},
		activeTried:  map[string]bool{},
		backoff:      firstBackoff,
	}
}
