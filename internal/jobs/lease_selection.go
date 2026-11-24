package jobs

import (
	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/policy"
)

type leaseSelection struct {
	request policy.LeaseRequest
	node    policy.Node
	now     float64
	cooling map[string]bool
	capped  map[string]bool
	recent  map[string]int
}

func (req *leaseSelection) pick(candidates []policy.Candidate, cfg *config.Config) (policy.Candidate, bool) {
	switch {
	case req.request.Profile != "":
		return policy.PickOnDemand(candidates, cfg, req.recent)
	case req.request.Kind == kindTask:
		return policy.PickTask(candidates, &req.request, cfg, req.recent, req.cooling)
	case req.request.Kind == kindOpenRouter:
		return policy.PickRemote(candidates, cfg, req.recent, req.now, func(c *policy.Candidate) bool {
			_, routed := policy.RunnerRouteProfile(c, cfg, req.cooling, req.node)
			return routed
		})
	default:
		return policy.PickJob(candidates, &req.request, cfg, req.recent, req.now)
	}
}
