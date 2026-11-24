package jobs

import (
	"context"
	"maps"

	"github.com/syntopica/worker/internal/store"
)

// finishFailed closes a job with a terminal control result ("failed" or
// "expired") that carries the code and the allowlisted extra fields.
func finishFailed(
	ctx context.Context, c store.Conn, job *jobRow, code, control string, extra map[string]any, now float64,
) error {
	err := run(ctx, c, "finish_failed", params{keyID: job.ID, keyState: control, keyCode: code, keyNow: now})
	if err != nil {
		return err
	}
	detail := map[string]any{keyError: code}
	maps.Copy(detail, extra)

	return addControlResult(ctx, c, job, control, detail, now)
}
