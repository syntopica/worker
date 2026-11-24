package policy

import "github.com/syntopica/worker/internal/config"

// fairest is the index of the pick: highest priority, then the lowest
// weighted share of recent attempts, then the oldest; -1 when there is none.
func fairest(candidates []Candidate, cfg *config.Config, recent map[string]int) int {
	best, bestShare := -1, 0.0
	for index := range candidates {
		c := &candidates[index]
		share := float64(recent[c.Queue]) / float64(max(cfg.Queues[c.Queue].Weight, 1))
		if best >= 0 {
			top := &candidates[best]
			if c.Priority < top.Priority {
				continue
			}
			if c.Priority == top.Priority && (share > bestShare || (share == bestShare && c.Created >= top.Created)) {
				continue
			}
		}
		best, bestShare = index, share
	}
	return best
}
