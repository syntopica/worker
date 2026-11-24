package policy

import (
	"slices"
	"strings"

	"github.com/syntopica/worker/internal/config"
)

// RunnerCandidates is what the task loop may lease: tasks as they are, and
// inference routed to a runner, carrying its route's profile in Model.
//
// An inference job is dropped while younger than its route's After or when
// the route is unusable. A shadow member pinned to a runner profile runs
// only under that profile; other pins are dropped.
func RunnerCandidates(
	candidates []Candidate, cfg *config.Config, cooling map[string]bool, node Node, now float64,
) []Candidate {
	var kept []Candidate
	for index := range candidates {
		c := &candidates[index]
		if c.Kind != kindInference && c.Pin == "" {
			kept = append(kept, *c)
			continue
		}
		profile, usable := pinnedRunnerProfile(c, cfg, cooling, node.Trust)
		if c.Pin == "" {
			route := cfg.Queues[c.Queue].Runner
			if route == nil || now-c.Created < route.After.Seconds() {
				continue
			}
			profile, usable = RunnerRouteProfile(c, cfg, cooling, node)
		}
		if usable {
			kept = append(kept, c.under(profile))
		}
	}
	return kept
}

// pinnedRunnerProfile is a runner pin's profile, or a judge's own, when
// usable now: configured, automatic, rested, and the job's class allowed by
// the profile and for runner executors, checked at lease time so a
// configuration change after the group opened still applies.
func pinnedRunnerProfile(c *Candidate, cfg *config.Config, cooling map[string]bool, trust string) (string, bool) {
	name := c.Model
	if c.Pin != pinJudge {
		kind, pinned, _ := strings.Cut(c.Pin, ":")
		if kind != kindRunner {
			return "", false
		}
		name = pinned
	}
	profile, known := cfg.Profiles[name]
	if !known || profile.OnDemand || ProfileResting(&profile, cooling) || !slices.Contains(profile.Privacy, c.Privacy) {
		return "", false
	}
	return name, PrivacyAllows(cfg, c.Privacy, kindRunner, trust)
}
