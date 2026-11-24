package policy

import "github.com/syntopica/worker/internal/config"

// RunnableProfile is the profile a queued task would run under on this node
// right now: its own, else its queue's first usable fallback, else none.
//
// Usable means configured, not on demand, allowed on the node and with a
// runner that is not resting after a quota wall. A pinned job never runs
// under a fallback.
func RunnableProfile(c *Candidate, node string, cfg *config.Config, cooling map[string]bool) (string, bool) {
	queue, found := cfg.Queues[c.Queue]
	if !found {
		return "", false
	}
	order := []string{c.Model}
	if c.Pin == "" {
		order = append(order, queue.Fallbacks[c.Model]...)
	}
	for _, name := range order {
		profile, known := cfg.Profiles[name]
		if known && !profile.OnDemand && !ProfileResting(&profile, cooling) && onNode(&profile, node) {
			return name, true
		}
	}
	return "", false
}
