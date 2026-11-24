package policy

import (
	"slices"

	"github.com/syntopica/worker/internal/config"
)

// OpenRouterMayTake reports whether a queue's OpenRouter route could take an
// inference job: the route maps the job's model, its class may go to
// OpenRouter, and the job does not exclude it.
func OpenRouterMayTake(c *Candidate, cfg *config.Config, trust string) bool {
	route := cfg.Queues[c.Queue].OpenRouter
	if route == nil || slices.Contains(c.ExcludeProviders, kindOpenRouter) {
		return false
	}
	_, mapped := route.Models[c.Model]
	return mapped && PrivacyAllows(cfg, c.Privacy, kindOpenRouter, trust)
}
