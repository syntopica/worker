package jobs

import (
	"context"
	"encoding/json"

	"github.com/syntopica/worker/internal/store"
)

func readJobResults(ctx context.Context, s *store.Store, id string) ([]map[string]any, error) {
	rows, err := queryMaps(ctx, s, "read_job_results", params{keyID: id})
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		for _, key := range []string{"detail", "executor", "usage"} {
			raw := text(row[key])
			if raw == "" {
				row[key] = nil
				continue
			}
			if !json.Valid([]byte(raw)) {
				return nil, errInvalidResult
			}
			row[key] = json.RawMessage(raw)
		}
	}
	return rows, nil
}
