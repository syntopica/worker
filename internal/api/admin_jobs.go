package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/syntopica/worker/internal/fault"
	"github.com/syntopica/worker/internal/jobs"
)

func adminJobs(ctx context.Context, r *request) (httpStatus int, payload any, failure error) {
	wanted, err := adminFilter(r)
	if err != nil {
		return 0, nil, err
	}
	rows, err := jobs.ListAdminJobs(ctx, r.store, wanted)
	return http.StatusOK, rows, err
}

func adminFilter(r *request) (jobs.Filter, error) {
	queue := r.query.optional("queue")
	if queue != nil {
		if _, exists := r.cfg.Queues[*queue]; !exists {
			return jobs.Filter{}, fault.New(http.StatusBadRequest, "unknown_queue")
		}
	}
	var states []string
	if state := r.query.optional("state"); state != nil {
		states = strings.Split(*state, ",")
		for _, name := range states {
			if !jobs.KnownState(name) {
				return jobs.Filter{}, fault.New(http.StatusBadRequest, "unknown_state")
			}
		}
	}
	limit, err := r.query.integer("limit", defaultLimit)
	if err != nil {
		return jobs.Filter{}, err
	}
	return jobs.Filter{
		Queue: queue, States: states, Producer: r.query.optional("producer"),
		Before: r.query.optional("before"), Limit: int(min(max(limit, 1), maxLimit)),
	}, nil
}
