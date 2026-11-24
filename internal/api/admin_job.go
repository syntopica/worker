package api

import (
	"context"
	"net/http"

	"github.com/syntopica/worker/internal/jobs"
)

func adminJob(ctx context.Context, r *request) (httpStatus int, payload any, failure error) {
	job, err := jobs.ReadAdminJob(ctx, r.store, r.id)
	return http.StatusOK, job, err
}
