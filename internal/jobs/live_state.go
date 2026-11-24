package jobs

// liveState reports whether a job in this state holds a lease.
func liveState(state string) bool {
	return state == stateLeased || state == stateRunning || state == stateDraining
}
