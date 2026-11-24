package api

import (
	"context"
	"net/http"

	"github.com/syntopica/worker/internal/jobs"
)

const maxAuditDays = 365

func adminAudit(ctx context.Context, r *request) (httpStatus int, payload any, failure error) {
	days, err := reportDays(r.query)
	if err != nil {
		return 0, nil, err
	}
	since := r.now - max(1, min(days, maxAuditDays))*daySeconds
	rows, err := jobs.ReadAudit(ctx, r.store, since)
	return http.StatusOK, map[string]any{"rows": rows}, err
}
