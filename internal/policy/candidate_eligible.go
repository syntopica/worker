package policy

import "github.com/syntopica/worker/internal/config"

// CandidateEligible reports whether one candidate may run on the requesting
// node right now (spec 7): queue run policy, incremental memory and parking.
// Privacy is filtered earlier.
func CandidateEligible(c *Candidate, req *LeaseRequest, cfg *config.Config) bool {
	queue, found := cfg.Queues[c.Queue]
	if !found {
		return false // a queue removed from the configuration runs nothing
	}
	if req.UserActive && queue.RunWhen != activeOK {
		return false
	}
	pin, found := cfg.Models[c.Model]
	if !found {
		return false
	}
	fits := pin.ColdGB <= cfg.Nodes[req.Node].MemoryBudgetGB
	if c.Model == req.ResidentModel {
		fits = pin.WarmGB <= req.FreeGB
	}
	return fits && (!c.Parked || req.CurrentIdle >= c.ParkedMinIdle)
}
