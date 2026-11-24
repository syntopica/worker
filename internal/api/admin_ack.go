package api

import (
	"context"
	"net/http"

	"github.com/syntopica/worker/internal/jobs"
)

func adminAck(ctx context.Context, r *request) (httpStatus int, payload any, failure error) {
	state, err := jobs.AdminAcknowledge(ctx, r.store, r.cfg, r.id, r.principal.Name, r.now)
	return http.StatusOK, map[string]string{"id": r.id, keyState: state}, err
}
