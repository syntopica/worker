package api

import (
	"context"
	"net/http"

	"github.com/syntopica/worker/internal/jobs"
)

const defaultHours = 24

func activity(ctx context.Context, r *request) (httpStatus int, payload any, failure error) {
	hours, err := r.query.number("hours", defaultHours)
	if err != nil {
		return 0, nil, err
	}
	data, err := jobs.ReadActivity(ctx, r.store, r.now, hours)
	return http.StatusOK, data, err
}
