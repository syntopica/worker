package policy

import (
	"slices"

	"github.com/syntopica/worker/internal/config"
)

// RunnerRouteProfile is the task profile a queue's runner route would run an
// inference job under now: the route's profile, else its first usable
// fallback, else none.
//
// Usable means configured, automatic, allowed on the node, not resting after
// a quota wall, and the job's class allowed by the profile; the class must
// also be allowed for runner executors. A profile whose runner the job
// excludes is skipped. A job a runner answered with nothing skips the rung:
// the next executor takes it.
func RunnerRouteProfile(c *Candidate, cfg *config.Config, cooling map[string]bool, node Node) (string, bool) {
	route := cfg.Queues[c.Queue].Runner
	if c.Error == "no_output" || route == nil || !PrivacyAllows(cfg, c.Privacy, kindRunner, node.Trust) {
		return "", false
	}
	for _, name := range append([]string{route.Profile}, route.Fallbacks...) {
		profile, known := cfg.Profiles[name]
		if !known || slices.Contains(c.ExcludeProviders, profile.Runner) {
			continue
		}
		if routeProfileUsable(&profile, c.Privacy, cooling, node.Name) {
			return name, true
		}
	}
	return "", false
}

// routeProfileUsable reports whether one profile can take a routed
// inference job right now.
func routeProfileUsable(profile *config.TaskProfile, privacy string, cooling map[string]bool, node string) bool {
	usable := !profile.OnDemand && !ProfileResting(profile, cooling) && onNode(profile, node)
	return usable && slices.Contains(profile.Privacy, privacy)
}
