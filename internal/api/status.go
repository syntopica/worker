package api

import (
	"context"
	"net/http"

	"github.com/syntopica/worker/internal/jobs"
)

func status(ctx context.Context, r *request) (httpStatus int, payload any, failure error) {
	data, err := jobs.ReadStatus(ctx, r.store, r.now)
	if err != nil {
		return 0, nil, err
	}
	data["openrouter"], err = jobs.OpenRouterHeadroom(ctx, r.store, r.cfg, r.now)
	return http.StatusOK, data, err
}
