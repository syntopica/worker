package policy

import (
	"strings"

	"github.com/syntopica/worker/internal/config"
)

// HeldForRemote reports whether the local model should leave an inference
// job to a remote executor: it is younger than the queue's LocalAfter while
// a remote route could take it.
//
// A shadow member is left alone unless it is pinned to the local model,
// which takes it at once. OpenRouter cannot take a capped queue's job today,
// so nothing is gained by holding it for that rung.
func HeldForRemote(c *Candidate, cfg *config.Config, cooling, capped map[string]bool, node Node, now float64) bool {
	if c.Pin != "" {
		return !strings.HasPrefix(c.Pin, kindOllama+":")
	}
	queue, found := cfg.Queues[c.Queue]
	if !found || now-c.Created >= queue.LocalAfter.Seconds() {
		return false
	}
	if _, routed := RunnerRouteProfile(c, cfg, cooling, node); routed {
		return true
	}
	return !capped[c.Queue] && OpenRouterMayTake(c, cfg, node.Trust)
}
