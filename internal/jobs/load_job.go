package jobs

import (
	"context"
	"database/sql"

	"github.com/syntopica/worker/internal/store"
)

// loadJob reads one job and reports whether it exists.
func loadJob(ctx context.Context, c store.Conn, id string) (*jobRow, error) {
	var job jobRow
	found, err := queryRow(ctx, c, "job", params{keyID: id},
		&job.ID, &job.Producer, &job.Queue, &job.Kind, &job.Key, &job.Priority, &job.Privacy, &job.Model,
		&job.State, &job.Attempts, &job.MaxAttempts, &job.Preemptions, &job.Parked, &job.ParkedRuns,
		&job.ParkedMinIdle, &job.Generation, &job.LeaseNode, &job.LeaseAttempt, &job.Error, &job.Tier,
		&job.ShadowOf, &job.Pin, &job.Exclude)

	if err == nil && !found {
		return nil, sql.ErrNoRows
	}
	return &job, err
}
