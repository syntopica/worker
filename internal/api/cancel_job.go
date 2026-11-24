package api

import (
	"context"
	"net/http"

	"github.com/syntopica/worker/internal/jobs"
)

func cancelJob(ctx context.Context, r *request) (httpStatus int, payload any, failure error) {
	state, err := jobs.CancelJob(ctx, r.store, &r.principal.Name, r.id, r.now)
	return http.StatusOK, map[string]string{keyState: state}, err
}
