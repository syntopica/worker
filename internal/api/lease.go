package api

import (
	"context"
	"net/http"

	"github.com/syntopica/worker/internal/jobs"
)

func lease(ctx context.Context, r *request) (httpStatus int, payload any, failure error) {
	b, err := decodeBody(r.raw, true)
	if err != nil {
		return 0, nil, err
	}
	if b["node"] != r.principal.Name {
		return 0, nil, forbidden()
	}
	if _, expireErr := jobs.ExpireLeases(ctx, r.store, r.now); expireErr != nil {
		return 0, nil, expireErr
	}
	offer, err := leaseRequest(r, b)
	if err != nil {
		return 0, nil, err
	}
	granted, err := jobs.LeaseJob(ctx, r.store, r.cfg, offer, r.now)
	if err != nil {
		return 0, nil, err
	}
	if granted == nil {
		return http.StatusNoContent, nil, nil
	}
	return http.StatusOK, granted, nil
}
