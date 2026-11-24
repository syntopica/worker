package api

import (
	"context"
	"net/http"

	"github.com/syntopica/worker/internal/jobs"
)

func getJob(ctx context.Context, r *request) (httpStatus int, payload any, failure error) {
	job, err := jobs.GetJob(ctx, r.store, r.cfg, r.principal.Name, r.id, r.now)
	return http.StatusOK, job, err
}
