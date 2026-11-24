package policy

import (
	"slices"

	"github.com/syntopica/worker/internal/config"
)

// LadderProviders holds the executor providers a queue's ladder could ever
// give a job: what the local, OpenRouter and runner rungs would report.
//
// A rung counts when the class may use its executor and the queue routes the
// job there; cooldowns and node trust are runtime state and are not read.
func LadderProviders(cfg *config.Config, queue, privacy, model string) map[string]bool {
	allowed := cfg.Privacy[privacy]
	policy := cfg.Queues[queue]
	providers := map[string]bool{}
	if slices.Contains(allowed, kindOllama) {
		providers[kindOllama] = true
	}
	if slices.Contains(allowed, kindOpenRouter) && policy.OpenRouter != nil && policy.OpenRouter.Models[model] != "" {
		providers[kindOpenRouter] = true
	}
	if slices.Contains(allowed, kindRunner) && policy.Runner != nil {
		addRunners(providers, cfg, policy.Runner, privacy)
	}
	return providers
}

// addRunners adds the runner of every automatic profile of a route that
// takes the class.
func addRunners(providers map[string]bool, cfg *config.Config, route *config.RunnerRoute, privacy string) {
	for _, name := range append([]string{route.Profile}, route.Fallbacks...) {
		profile, known := cfg.Profiles[name]
		if known && !profile.OnDemand && slices.Contains(profile.Privacy, privacy) {
			providers[profile.Runner] = true
		}
	}
}
