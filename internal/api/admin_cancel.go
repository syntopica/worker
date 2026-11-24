package api

import (
	"context"
	"net/http"

	"github.com/syntopica/worker/internal/jobs"
)

func adminCancel(ctx context.Context, r *request) (httpStatus int, payload any, failure error) {
	state, err := jobs.AdminCancelJob(ctx, r.store, r.id, r.principal.Name, r.now)
	return http.StatusOK, map[string]string{"id": r.id, keyState: state}, err
}
