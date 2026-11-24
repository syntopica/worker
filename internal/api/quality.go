package api

import (
	"context"
	"net/http"

	"github.com/syntopica/worker/internal/jobs"
	"github.com/syntopica/worker/internal/store"
)

func quality(ctx context.Context, r *request) (httpStatus int, payload any, failure error) {
	days, err := reportDays(r.query)
	if err != nil {
		return 0, nil, err
	}
	since := r.now - days*daySeconds
	counts := make(map[string]any)
	for _, section := range []struct {
		key  string
		read func(context.Context, *store.Store, float64) ([]map[string]any, error)
	}{
		{"attempts", jobs.ReadAttemptQuality},
		{"ratings", jobs.ReadResultRatings},
		{"judged", jobs.ReadJudgements},
	} {
		rows, err := section.read(ctx, r.store, since)
		if err != nil {
			return 0, nil, err
		}
		counts[section.key] = rows
	}
	return http.StatusOK, counts, nil
}
