package tasks

// reason is the allowlisted error code of an attempt that did not succeed.
type reason string

const (
	unknownProfile    reason = "unknown_profile"
	privacyRefused    reason = "privacy_refused"
	nodeShutdown      reason = "node_shutdown"
	nodeError         reason = "node_error"
	inputsNotAllowed  reason = "inputs_not_allowed"
	inputMissing      reason = "input_missing"
	inputDenied       reason = "input_denied"
	inputTooLarge     reason = "input_too_large"
	runnerFailed      reason = "runner_failed"
	runnerAuth        reason = "runner_auth"
	rateLimited       reason = "rate_limited"
	runnerUnavailable reason = "runner_unavailable"
	timedOut          reason = "timeout"
	quotaWall         reason = "quota_wall"
	noOutput          reason = "no_output"
)
