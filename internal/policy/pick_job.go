package policy

import "github.com/syntopica/worker/internal/config"

// PickJob chooses the next job (spec 7): the resident model's, unless
// another model's job is overdue, then priority, weighted share and age.
// recent counts attempts started per queue in the last hour.
func PickJob(
	candidates []Candidate, req *LeaseRequest, cfg *config.Config, recent map[string]int, now float64,
) (Candidate, bool) {
	var eligible, resident, overdue []Candidate
	for index := range candidates {
		c := &candidates[index]
		if !CandidateEligible(c, req, cfg) {
			continue
		}
		eligible = append(eligible, *c)
		switch {
		case c.Model == req.ResidentModel:
			resident = append(resident, *c)
		case now-c.Created > cfg.Queues[c.Queue].MaxModelAge.Seconds():
			overdue = append(overdue, *c)
		}
	}
	pool := eligible
	if len(overdue) > 0 {
		pool = overdue
	} else if len(resident) > 0 {
		pool = resident
	}
	chosen := fairest(pool, cfg, recent)
	if chosen < 0 {
		return Candidate{}, false
	}
	return pool[chosen], true
}
