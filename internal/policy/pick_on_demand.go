package policy

import "github.com/syntopica/worker/internal/config"

// PickOnDemand chooses the next job for an on-demand profile among
// candidates already filtered for it: priority, weighted share, age.
//
// No run policy: the owner asked for this run, so an active user holds
// nothing back.
func PickOnDemand(candidates []Candidate, cfg *config.Config, recent map[string]int) (Candidate, bool) {
	chosen := fairest(candidates, cfg, recent)
	if chosen < 0 {
		return Candidate{}, false
	}
	return candidates[chosen], true
}
