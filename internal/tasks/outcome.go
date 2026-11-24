package tasks

// outcome is how an attempt ended, as the contract names it.
type outcome string

const (
	succeeded outcome = "succeeded"
	failed    outcome = "failed"
	preempted outcome = "preempted"
)
