package api

import (
	"context"
	"net/http"

	"github.com/syntopica/worker/internal/jobs"
)

func complete(ctx context.Context, r *request) (httpStatus int, payload any, failure error) {
	b, err := decodeBody(r.raw, true)
	if err != nil {
		return 0, nil, err
	}
	report, err := completionReport(b)
	if err != nil {
		return 0, nil, err
	}
	generation, err := b.generation()
	if err != nil {
		return 0, nil, err
	}
	state, err := jobs.CompleteAttempt(ctx, r.store, r.cfg, r.id, generation, r.principal.Name, report, r.now)
	return http.StatusOK, map[string]string{keyState: state}, err
}

func completionReport(b body) (*jobs.CompletionReport, error) {
	value, found := b["outcome"]
	if !found {
		return nil, badRequest()
	}
	wall, err := b.number("wall_s", 0)
	if err != nil {
		return nil, err
	}
	outcome, _ := value.(string)
	errorCode, _ := b["error_code"].(string)
	executor, _ := b["executor"].(map[string]any)
	usage, _ := b["usage"].(map[string]any)
	return &jobs.CompletionReport{
		Outcome: outcome, ErrorCode: errorCode, WallS: wall, Output: b["output"], Usage: usage, Executor: executor,
	}, nil
}
