package node

// outcome is how an attempt ended, as the node sees it.
type outcome string

const (
	succeeded   outcome = "succeeded"
	failed      outcome = "failed"
	preempted   outcome = "preempted"
	fenced      outcome = "fenced"
	drainFailed outcome = "drain_failed"
)

// ran reports whether the attempt ran on the model, and so loaded it.
func (o outcome) ran() bool {
	return o == succeeded || o == preempted || o == fenced || o == drainFailed
}
