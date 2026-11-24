package jobs

import (
	"regexp"
	"slices"
)

var httpCode = regexp.MustCompile(`^http_\d{3}$`)

// knownErrorCode holds a reported error code to the fixed vocabulary
// (spec 9): the code itself when known, else executor_error, so free text
// never reaches storage. An empty code means none was reported.
func knownErrorCode(code string) string {
	known := []string{
		"bad_response", stateCancelled, "drain_failed", codeExecutorError, "host_state_unreadable",
		"input_denied", "input_missing", "input_too_large", "inputs_not_allowed", "lease_lost",
		"memory_pressure", "no_output", "node_error", "node_shutdown", "on_battery", "privacy_refused",
		codeQuotaWall, "rate_limited", "runner_auth", "runner_failed", "runner_unavailable",
		codeSchemaViolation, "timeout", "transport_error", "unknown_model", "unknown_profile", "user_active",
	}
	if code == "" || slices.Contains(known, code) || httpCode.MatchString(code) {
		return code
	}

	return codeExecutorError
}
