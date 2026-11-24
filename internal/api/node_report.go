package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/syntopica/worker/internal/jobs"
)

func nodeReport(ctx context.Context, r *request) (httpStatus int, payload any, failure error) {
	if r.id != r.principal.Name {
		return 0, nil, forbidden()
	}
	report, err := decodeBody(r.raw, false)
	if err != nil {
		return 0, nil, err
	}
	encoded, err := json.Marshal(report)
	if err != nil {
		return 0, nil, fmt.Errorf("encode node metadata: %w", err)
	}
	err = jobs.SaveNodeReport(ctx, r.store, r.id, string(encoded), r.now)
	return http.StatusOK, map[string]bool{"ok": true}, err
}
