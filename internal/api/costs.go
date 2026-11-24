package api

import (
	"context"
	"net/http"

	"github.com/syntopica/worker/internal/jobs"
)

func costs(ctx context.Context, r *request) (httpStatus int, payload any, failure error) {
	days, err := reportDays(r.query)
	if err != nil {
		return 0, nil, err
	}
	rows, err := jobs.ReadCosts(ctx, r.store, r.now-days*daySeconds)
	return http.StatusOK, map[string]any{"rows": rows}, err
}
