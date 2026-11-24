package tasks

// ending is how the wait for a runner process ended.
type ending int

const (
	// exited is a runner that ended on its own.
	exited ending = iota
	// fencedOut is a runner killed because its lease is someone else's now.
	fencedOut
	// overran is a runner killed at its profile's timeout.
	overran
	// stopped is a runner killed because the loop was told to stop.
	stopped
)
