package policy

import "github.com/syntopica/worker/internal/config"

// PickTask chooses the next task: priority, then weighted share across
// queues, then age. The pick carries the profile it will run under in
// Model: its own, or a queue fallback while its runner rests.
func PickTask(
	candidates []Candidate, req *LeaseRequest, cfg *config.Config, recent map[string]int, cooling map[string]bool,
) (Candidate, bool) {
	var eligible []Candidate
	for index := range candidates {
		if TaskEligible(&candidates[index], req, cfg, cooling) {
			eligible = append(eligible, candidates[index])
		}
	}
	chosen := fairest(eligible, cfg, recent)
	if chosen < 0 {
		return Candidate{}, false
	}
	pick := eligible[chosen]
	if profile, runnable := RunnableProfile(&pick, req.Node, cfg, cooling); runnable {
		pick.Model = profile
	}
	return pick, true
}
