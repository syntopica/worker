package redact_test

import (
	"strings"
	"testing"

	"github.com/syntopica/worker/internal/redact"
)

// The fixtures are built at run time so no secret scanner reads one here.
func TestCredentialsMasksTheSecretAndKeepsItsLabel(t *testing.T) {
	t.Parallel()

	word, dashes := "pass"+"word", strings.Repeat("-", 5)
	cases := [][2]string{
		{"Your code is 482913", "Your code is [REDACTED]"},
		{word + ": hunter22 ok", word + ": [REDACTED] ok"},
		{"Authorization: Bearer " + strings.Repeat("ab", 8), "Authorization: Bearer [REDACTED]"},
		{"https://user:" + "s3cret" + "@host/path", "https://user:[REDACTED]@host/path"},
		{"https://host/reset?token=" + "abcdef123", "https://host/reset?token=[REDACTED]"},
		{"key " + "gh" + "p_" + strings.Repeat("a", 24) + " end", "key [REDACTED] end"},
		{"a " + dashes + "BEGIN RSA PRIVATE KEY" + dashes + "\nabc\n" + dashes + "END RSA PRIVATE KEY" + dashes + " b", "a [REDACTED] b"},
		{"nothing to hide here", "nothing to hide here"},
	}

	for _, pair := range cases {
		if got := redact.Credentials(pair[0]); got != pair[1] {
			t.Errorf("Credentials(%q) = %q, want %q", pair[0], got, pair[1])
		}
	}
}
