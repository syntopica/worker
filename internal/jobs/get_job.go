package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/store"
)

// GetJob returns the current state and latest unacknowledged result of a producer's job.
func GetJob(ctx context.Context, s *store.Store, cfg *config.Config, producer, id string, now float64) (JobStatus, error) {
	job, err := ownedJob(ctx, s, &producer, id)
	if err != nil {
		return JobStatus{}, err
	}
	status := JobStatus{ID: id, State: job.State}
	if job.Error != "" {
		status.Error = &job.Error
	}
	var result JobResult
	found, err := queryRow(ctx, s, "latest_result", params{keyProducer: producer, keyID: id}, result.dest()...)
	if err != nil {
		return JobStatus{}, err
	}
	if found {
		status.Result = &result
	}
	status.CoolingUntil, err = jobCoolingUntil(ctx, s, cfg, job, now)
	return status, err
}
