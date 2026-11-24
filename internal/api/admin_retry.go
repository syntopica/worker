package api

import (
	"context"

	"github.com/syntopica/worker/internal/jobs"
)

func adminRetry(ctx context.Context, r *request) (httpStatus int, payload any, failure error) {
	submitted, err := jobs.RetryJob(ctx, r.store, r.cfg, r.id, r.principal.Name, r.now)
	if err != nil {
		return 0, nil, err
	}
	state, err := jobs.JobState(ctx, r.store, submitted.ID)
	return createdStatus(submitted.Created), map[string]string{"id": submitted.ID, keyState: state, "retry_of": r.id}, err
}
