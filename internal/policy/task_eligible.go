package policy

import "github.com/syntopica/worker/internal/config"

// TaskEligible reports whether one task candidate may run on the requesting
// node right now: queue run policy, then a usable profile.
//
// No memory check: the runner is a CLI talking to a remote model. A profile
// removed from the configuration runs nothing, like a removed queue.
func TaskEligible(c *Candidate, req *LeaseRequest, cfg *config.Config, cooling map[string]bool) bool {
	queue, found := cfg.Queues[c.Queue]
	if _, known := cfg.Profiles[c.Model]; !found || !known {
		return false
	}
	if req.UserActive && queue.RunWhen != activeOK {
		return false
	}
	_, runnable := RunnableProfile(c, req.Node, cfg, cooling)
	return runnable
}
