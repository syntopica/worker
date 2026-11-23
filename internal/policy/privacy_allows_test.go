package policy_test

import (
	"testing"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/policy"
)

func TestPrivacyAllows(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Privacy: map[string][]string{"mail": {"ollama"}, "public": {"ollama", "openrouter"}},
		Trust:   map[string][]string{"mail": {"owner"}, "public": {"owner", "guest"}},
	}

	cases := []struct {
		privacy, executor, trust string
		want                     bool
	}{
		{"mail", "ollama", "owner", true},
		{"mail", "openrouter", "owner", false},
		{"mail", "ollama", "guest", false},
		{"public", "openrouter", "guest", true},
		{"", "ollama", "owner", false},
		{"secret", "ollama", "owner", false},
	}

	for _, c := range cases {
		if got := policy.PrivacyAllows(cfg, c.privacy, c.executor, c.trust); got != c.want {
			t.Errorf("PrivacyAllows(%q, %q, %q) = %v, want %v", c.privacy, c.executor, c.trust, got, c.want)
		}
	}
}
