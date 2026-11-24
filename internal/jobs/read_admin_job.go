package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/store"
)

// ReadAdminJob returns metadata, attempts, results and payload presence for a job.
func ReadAdminJob(ctx context.Context, s *store.Store, id string) (map[string]any, error) {
	values := adminParams(1)
	values[keyID] = id
	rows, err := adminJobRows(ctx, s, values)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, notFound()
	}
	return addAdminDetails(ctx, s, id, rows[0])
}

func addAdminDetails(ctx context.Context, s *store.Store, id string, row map[string]any) (map[string]any, error) {
	values := params{keyID: id}
	attempts, err := queryMaps(ctx, s, "read_job_attempts", values)
	if err != nil {
		return nil, err
	}
	results, err := readJobResults(ctx, s, id)
	if err != nil {
		return nil, err
	}
	var input, output bool
	if _, err := queryRow(ctx, s, "admin_payload_presence", values, &input, &output); err != nil {
		return nil, err
	}
	row["attempt_details"], row["results"] = attempts, results
	row["has_input"], row["has_output"] = input, output
	return row, nil
}
