package jobs

import (
	"slices"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/policy"
)

// leaseCandidates narrows the ready jobs to the requested executor rung.
func leaseCandidates(candidates []policy.Candidate, cfg *config.Config, req *leaseSelection) []policy.Candidate {
	switch {
	case req.request.Profile != "":
		return policy.OnDemandCandidates(candidates, cfg, req.cooling, req.node, req.request.Profile, req.request.OverridePins)
	case req.request.Kind == kindTask:
		return policy.RunnerCandidates(candidates, cfg, req.cooling, req.node, req.now)
	case req.request.Kind == kindOpenRouter:
		return slices.DeleteFunc(candidates, func(c policy.Candidate) bool {
			return req.capped[c.Queue] || slices.Contains(c.ExcludeProviders, kindOpenRouter)
		})
	default:
		return slices.DeleteFunc(candidates, func(c policy.Candidate) bool {
			return slices.Contains(c.ExcludeProviders, executorOllama) ||
				policy.HeldForRemote(&c, cfg, req.cooling, req.capped, req.node, req.now)
		})
	}
}
