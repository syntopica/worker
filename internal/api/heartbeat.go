package api

import (
	"context"
	"net/http"

	"github.com/syntopica/worker/internal/jobs"
)

func heartbeat(ctx context.Context, r *request) (httpStatus int, payload any, failure error) {
	b, err := decodeBody(r.raw, true)
	if err != nil {
		return 0, nil, err
	}
	generation, err := b.generation()
	if err != nil {
		return 0, nil, err
	}
	err = jobs.HeartbeatAttempt(ctx, r.store, r.id, generation, r.principal.Name, truthy(b["draining"]), r.now)
	return http.StatusOK, map[string]bool{"ok": true}, err
}
