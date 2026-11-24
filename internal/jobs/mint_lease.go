package jobs

import (
	"context"
	"encoding/json"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/contract"
	"github.com/syntopica/worker/internal/policy"
	"github.com/syntopica/worker/internal/store"
)

// mintLease persists the generation and attempt before a lease can be sent.
func mintLease(
	ctx context.Context, tx store.Conn, cfg *config.Config, req *leaseSelection, c *policy.Candidate, input json.RawMessage,
) (*contract.Lease, error) {
	lease := &contract.Lease{
		JobID: c.JobID, AttemptID: newID(), Model: c.Model, Input: input,
		RunWhen: cfg.Queues[c.Queue].RunWhen, TTLSeconds: leaseTTL,
		Kind: req.request.Kind, Queue: c.Queue, Privacy: c.Privacy, Pinned: c.Pin != "",
	}
	if req.request.OverridePins && c.Kind == kindInference && lease.Pinned {
		err := recordAudit(ctx, tx, c.JobID, c.Privacy, req.node.Name, "override_pin", req.now)
		if err != nil {
			return nil, err
		}
	}
	values := params{
		keyID: c.JobID, keyAttempt: lease.AttemptID, keyNode: req.node.Name, keyNow: req.now,
		keyModel: c.Model, keyProfile: nullIfEmpty(req.request.Profile), keyExpires: req.now + leaseTTL,
	}
	if _, err := queryRow(ctx, tx, "mint_lease", values, &lease.Generation); err != nil {
		return nil, err
	}
	var profile any
	if req.request.Kind == kindTask {
		profile = c.Model
	}
	if err := run(ctx, tx, "insert_attempt", params{
		keyID: c.JobID, keyAttempt: lease.AttemptID, keyNode: req.node.Name, keyNow: req.now,
		keyGeneration: lease.Generation, keyProfile: profile,
	}); err != nil {
		return nil, err
	}
	return lease, nil
}
