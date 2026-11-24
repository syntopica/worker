package jobs

import "slices"

// KnownState reports whether a name is a job state (spec 6).
func KnownState(state string) bool {
	return slices.Contains([]string{
		stateQueued, stateLeased, stateRunning, stateDraining, stateSplitRequested, stateSuperseded,
		"needs_reconciliation", "reconciled", stateSucceeded, stateFailed, stateExpired, stateCancelled,
		stateUnackedExpired,
	}, state)
}
