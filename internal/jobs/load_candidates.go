package jobs

import (
	"context"
	"database/sql"
	"maps"
	"slices"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/policy"
	"github.com/syntopica/worker/internal/store"
)

// loadCandidates applies queue and privacy limits before ranking, so a
// backlog the node cannot see cannot hide work it may run.
func loadCandidates(
	ctx context.Context, tx store.Conn, cfg *config.Config, req *policy.LeaseRequest, trust string, now float64,
) ([]policy.Candidate, error) {
	executor, extraKind := executorOllama, kindInference
	switch req.Kind {
	case kindTask:
		executor, extraKind = executorRunner, kindTask
	case kindOpenRouter:
		executor = kindOpenRouter
	}

	classes := []string{}
	for class := range cfg.Privacy {
		if policy.PrivacyAllows(cfg, class, executor, trust) {
			classes = append(classes, class)
		}
	}
	if len(classes) == 0 || len(cfg.Queues) == 0 {
		return nil, nil
	}
	allowed, err := encodeValue(classes)
	if err != nil {
		return nil, err
	}
	queues, err := encodeValue(slices.Collect(maps.Keys(cfg.Queues)))
	if err != nil {
		return nil, err
	}
	return queryAll(ctx, tx, "lease_candidates", params{
		"extra_kind": extraKind, "created_before": req.CreatedBefore, keyNow: now,
		"classes": allowed, "queues": queues,
	}, scanCandidate)
}

func scanCandidate(rows *sql.Rows) (policy.Candidate, error) {
	var c policy.Candidate
	var excluded string
	err := scanInto(rows, &c.JobID, &c.Queue, &c.Model, &c.Priority, &c.Privacy, &c.Created,
		&c.Parked, &c.ParkedMinIdle, &c.Kind, &c.Pin, &c.Error, &excluded)
	if err != nil {
		return c, err
	}
	c.ExcludeProviders, err = excludedProviders(excluded)
	return c, err
}
