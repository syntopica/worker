package tasks

import (
	"strings"
	"testing"
)

func TestTheTailOfARunnersMessagesNamesTheFailure(t *testing.T) {
	t.Parallel()

	cases := map[string]reason{
		"ERROR: unexpected status 401 Unauthorized":    runnerAuth,
		"error: 429 Too Many Requests":                 rateLimited,
		"stream disconnected before completion":        runnerUnavailable,
		"HTTP 503 from upstream":                       runnerUnavailable,
		"panic: something else":                        runnerFailed,
		"":                                             runnerFailed,
		"401 unauthorized" + strings.Repeat("x", 5000): runnerFailed,
	}

	for messages, want := range cases {
		if got := failureCode(messages); got != want {
			t.Errorf("%.40q: %s", messages, got)
		}
	}
}
