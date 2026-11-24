package remote

// reason is the allowlisted error code of an attempt that did not succeed.
type reason string

const (
	privacyRefused reason = "privacy_refused"
	nodeShutdown   reason = "node_shutdown"
	nodeError      reason = "node_error"
	rateLimited    reason = "rate_limited"
	transportError reason = "transport_error"
	badResponse    reason = "bad_response"
)
