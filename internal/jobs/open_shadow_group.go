package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/store"
)

// openShadowGroup preserves a copy of the original answer for the judge and
// enqueues the allowed comparison targets within the queue's pending cap.
func openShadowGroup(
	ctx context.Context, tx store.Conn, cfg *config.Config, job *jobRow, input map[string]any,
	report *CompletionReport, now float64,
) error {
	shadow := cfg.Queues[job.Queue].Shadow
	if shadow == nil || job.Kind != kindInference || job.ShadowOf != "" || !shadowSampled(job.ID, shadow.Rate) {
		return nil
	}
	if !targetAllowed(cfg, job.Privacy, executorRunner+":"+shadow.Judge) {
		return nil
	}
	targets := shadowTargets(cfg, job.Privacy, shadow.Targets, report.Executor)
	var pending int
	_, err := queryRow(ctx, tx, "shadow_pending", params{keyQueue: job.Queue}, &pending)
	if err != nil || len(targets) == 0 || pending+len(targets) > shadow.MaxPending {
		return err
	}
	return insertShadowGroup(ctx, tx, job, input, targets, report, now)
}
