package api

import (
	"context"
	"net/http"

	"github.com/syntopica/worker/internal/jobs"
)

func adminContent(ctx context.Context, r *request) (httpStatus int, payload any, failure error) {
	data, err := jobs.ReadAdminContent(ctx, r.store, r.id, r.reveal, r.principal.Name, r.now)
	return http.StatusOK, data, err
}
