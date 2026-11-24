package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/store"
)

func acceptOutput(
	ctx context.Context, tx store.Conn, cfg *config.Config, job *jobRow, report *CompletionReport, now float64,
) (string, error) {
	input, err := loadJobInput(ctx, tx, job.ID)
	if err != nil {
		return "", err
	}
	if input == nil {
		return stateFailed, failPayloadLost(ctx, tx, job.ID, now)
	}
	if path, valid := checkOutputSchema(input, report.Output); !valid {
		return rejectOutput(ctx, tx, job, path, now)
	}
	if err := addOutput(ctx, tx, job, report, now); err != nil {
		return "", err
	}
	if err := run(ctx, tx, "accept_output", params{keyID: job.ID, keyNow: now}); err != nil {
		return "", err
	}
	return stateSucceeded, openShadowGroup(ctx, tx, cfg, job, input, report, now)
}

func rejectOutput(ctx context.Context, tx store.Conn, job *jobRow, path string, now float64) (string, error) {
	if err := run(ctx, tx, "schema_violation", params{keyAttempt: job.LeaseAttempt}); err != nil {
		return "", err
	}
	detail := map[string]any{}
	if path != "" {
		detail["schema_path"] = path
	}
	return failAttempt(ctx, tx, job, codeSchemaViolation, detail, now)
}
