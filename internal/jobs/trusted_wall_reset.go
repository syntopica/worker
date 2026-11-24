package jobs

const maxWallReset = 14 * daySeconds

// trustedWallReset accepts only numeric durations within the node's bound.
func trustedWallReset(value any) float64 {
	var seconds float64
	switch typed := value.(type) {
	case float64:
		seconds = typed
	case int64:
		seconds = float64(typed)
	}
	if seconds > 0 && seconds <= maxWallReset {
		return seconds
	}
	return 0
}
