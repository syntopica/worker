package remote

// outcome is how an attempt ended, as the loop sees it.
type outcome string

const (
	succeeded outcome = "succeeded"
	failed    outcome = "failed"
	preempted outcome = "preempted"
	fenced    outcome = "fenced"
)
