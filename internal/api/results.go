package api

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/syntopica/worker/internal/fault"
	"github.com/syntopica/worker/internal/jobs"
)

const (
	maxResultWait = 30
	pollInterval  = 500 * time.Millisecond
)

func results(ctx context.Context, r *request) (httpStatus int, payload any, failure error) {
	queue := r.query["queue"]
	if !slices.Contains(r.cfg.Producers[r.principal.Name], queue) {
		return 0, nil, fault.New(http.StatusForbidden, "queue_not_granted")
	}
	after, err := r.query.integer("after", 0)
	if err != nil {
		return 0, nil, err
	}
	limit, err := r.query.integer("limit", defaultLimit)
	if err != nil {
		return 0, nil, err
	}
	wait, err := r.query.number("wait", 0)
	if err != nil {
		return 0, nil, err
	}
	rows, err := pollResults(ctx, r, queue, after, int(min(max(limit, 1), maxLimit)),
		time.Duration(min(max(wait, 0), maxResultWait)*float64(time.Second)))
	return http.StatusOK, map[string]any{"results": rows}, err
}

func pollResults(ctx context.Context, r *request, queue string, after int64, limit int, wait time.Duration) ([]jobs.JobResult, error) {
	deadline := time.Now().Add(wait)
	for {
		if _, err := jobs.ExpireLeases(ctx, r.store, epochSeconds()); err != nil {
			return nil, err
		}
		rows, err := jobs.ListResults(ctx, r.store, r.principal.Name, queue, after, limit)
		left := time.Until(deadline)
		if err != nil || len(rows) > 0 || left <= 0 {
			return rows, err
		}
		timer := time.NewTimer(min(left, pollInterval))
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, fmt.Errorf("poll results: %w", ctx.Err())
		case <-timer.C:
		}
	}
}
