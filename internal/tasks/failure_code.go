package tasks

import "regexp"

// failureTail is how much of the end of a runner's messages is read.
const failureTail = 2048

// The signs in a runner's own messages that tell why it exited without an
// answer.
var (
	authFailure = regexp.MustCompile(`(?i)\b401\b|unauthori[sz]ed|not logged in|please (?:log|sign) in` +
		`|authentication (?:failed|required)|invalid api key`)
	rateFailure        = regexp.MustCompile(`(?i)\b429\b|rate.?limit|too many requests`)
	unavailableFailure = regexp.MustCompile(`(?i)\b50[234]\b|overloaded|service unavailable|temporarily unavailable` +
		`|stream disconnected|connection (?:reset|refused|closed)|timed out`)
)

// failureCode names why a runner failed from the end of its own messages:
// the first matching code, else runner_failed. The code is a label for the
// owner; it rests nothing, and the text itself is never kept.
func failureCode(messages string) reason {
	tail := messages[max(0, len(messages)-failureTail):]

	switch {
	case authFailure.MatchString(tail):
		return runnerAuth
	case rateFailure.MatchString(tail):
		return rateLimited
	case unavailableFailure.MatchString(tail):
		return runnerUnavailable
	}

	return runnerFailed
}
