package node

import (
	"context"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/contract"
	"github.com/syntopica/worker/internal/policy"
)

// runLeased runs one granted lease and reports whether the node should rest
// before its next one.
//
// The node checks the lease's privacy class itself: the coordinator already
// filtered by the same policy, and this catches the two disagreeing (a stale
// or edited configuration). A lease without a class is refused.
func (n *Node) runLeased(
	ctx context.Context, cfg *config.Config, node *config.NodePolicy, lease *contract.Lease, wasResident bool,
) bool {
	run := &attempt{
		node:     n,
		policy:   node,
		lease:    lease,
		backend:  n.Backend(node.OllamaURL),
		executor: contract.Executor{Node: node.Name, Provider: provider, Model: lease.Model},
	}

	pin, known := cfg.Models[lease.Model]

	refusal := reason("")

	switch {
	case !policy.PrivacyAllows(cfg, lease.Privacy, provider, node.Trust):
		refusal = privacyRefused
	case !known:
		refusal = unknownModel
	}

	if refusal != "" {
		run.complete(ctx, failed, refusal, nil, 0)

		return false
	}

	run.pin = &pin

	how, why, err := run.run(ctx)
	if err != nil {
		n.logf("worker: attempt failed: %s\n", err)
		run.complete(ctx, failed, nodeError, nil, 0)

		how, why = failed, nodeError
	}

	if ctx.Err() != nil {
		return false
	}

	if how != succeeded {
		// Outcome and code are fixed allowlisted words; the job id is opaque.
		n.logf("worker: attempt %s: %s job=%s\n", how, why, lease.JobID)
	}

	return settleAttempt(n.remembered(), lease.Model, wasResident, how, why, n.Now())
}
