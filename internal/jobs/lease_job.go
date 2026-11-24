package jobs

import (
	"context"
	"encoding/json"
	"net/http"
	"slices"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/contract"
	"github.com/syntopica/worker/internal/fault"
	"github.com/syntopica/worker/internal/policy"
	"github.com/syntopica/worker/internal/store"
)

const maxLostPerLease = 16

// LeaseJob returns one eligible job, or nil when the node has no work.
func LeaseJob(
	ctx context.Context, s *store.Store, cfg *config.Config, request *policy.LeaseRequest, now float64,
) (*contract.Lease, error) {
	node, known := cfg.Nodes[request.Node]
	if !known {
		return nil, fault.New(http.StatusForbidden, "unknown_node")
	}
	normalized := *request
	if normalized.Kind == "" {
		normalized.Kind = kindInference
	}
	req := &leaseSelection{request: normalized, node: policy.Node{Name: request.Node, Trust: node.Trust}, now: now}
	lease, err := write(ctx, s, func(tx store.Conn) (*contract.Lease, error) {
		return selectLease(ctx, tx, cfg, req)
	})
	if err != nil {
		return nil, err
	}
	return lease, nil
}

func selectLease(ctx context.Context, tx store.Conn, cfg *config.Config, req *leaseSelection) (*contract.Lease, error) {
	candidates, err := loadCandidates(ctx, tx, cfg, &req.request, req.node.Trust, req.now)
	if err != nil || len(candidates) == 0 {
		return nil, err
	}
	req.cooling, err = coolingRunners(ctx, tx, cfg, req.now)
	if err != nil {
		return nil, err
	}
	spent, err := openrouterSpentToday(ctx, tx, req.now)
	if err != nil {
		return nil, err
	}
	req.capped = policy.CappedQueues(cfg, spent)
	candidates = leaseCandidates(candidates, cfg, req)
	if profile, found := cfg.Profiles[req.request.Profile]; found && req.request.Profile != "" {
		candidates, err = onDemandPayloads(ctx, tx, candidates, &profile)
		if err != nil {
			return nil, err
		}
	}
	req.recent, err = queryCounts(ctx, tx, "recent_attempts", params{keySince: req.now - hourSeconds})
	if err != nil {
		return nil, err
	}
	return pickLease(ctx, tx, cfg, req, candidates)
}

func pickLease(
	ctx context.Context, tx store.Conn, cfg *config.Config, req *leaseSelection, candidates []policy.Candidate,
) (*contract.Lease, error) {
	var empty *contract.Lease
	for range maxLostPerLease {
		c, found := req.pick(candidates, cfg)
		if !found {
			break
		}
		input, err := loadJobInput(ctx, tx, c.JobID)
		if err != nil {
			return nil, err
		}
		if input != nil {
			raw, err := leaseInput(req.request.Kind, c.Kind, input)
			if err != nil {
				return nil, err
			}
			return mintLease(ctx, tx, cfg, req, &c, raw)
		}
		if err := failPayloadLost(ctx, tx, c.JobID, req.now); err != nil {
			return nil, err
		}
		candidates = slices.DeleteFunc(candidates, func(next policy.Candidate) bool { return next.JobID == c.JobID })
	}
	return empty, nil
}

func leaseInput(requestKind, jobKind string, input map[string]any) (json.RawMessage, error) {
	if requestKind == kindTask && jobKind == kindInference {
		task, err := inferenceTaskInput(input)
		if err != nil {
			return nil, err
		}
		input = task
	}
	raw, err := json.Marshal(input)
	if err != nil {
		return nil, errInvalidInput
	}
	return raw, nil
}
