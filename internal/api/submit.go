package api

import (
	"context"

	"github.com/syntopica/worker/internal/contract"
	"github.com/syntopica/worker/internal/jobs"
)

func submit(ctx context.Context, r *request) (httpStatus int, payload any, failure error) {
	id, created, err := jobs.SubmitJob(ctx, r.store, r.cfg, r.principal.Name, r.raw, r.now)
	return createdStatus(created), contract.Submitted{ID: id, Created: created}, err
}
