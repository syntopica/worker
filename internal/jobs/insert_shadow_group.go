package jobs

import (
	"context"
	"strings"

	"github.com/syntopica/worker/internal/store"
)

func insertShadowGroup(
	ctx context.Context, tx store.Conn, job *jobRow, input map[string]any, targets []string, report *CompletionReport, now float64,
) error {
	member := &shadowMember{group: job.ID, producer: producerShadow, kind: kindInference, maxAttempts: 1}
	for _, target := range targets {
		model := job.Model
		if local, ok := strings.CutPrefix(target, executorOllama+":"); ok {
			model = local
		}
		member.pin, member.model = target, model
		if _, err := insertShadowJob(ctx, tx, job, input, member, now); err != nil {
			return err
		}
	}
	member.pin, member.model = pinOriginal, job.Model
	original, err := insertShadowJob(ctx, tx, job, input, member, now)
	if err != nil {
		return err
	}
	if err := run(ctx, tx, "finish_original_shadow", params{keyID: original, keyNow: now}); err != nil {
		return err
	}
	originalJob := &jobRow{ID: original, Queue: job.Queue, Producer: producerShadow}
	return addOutput(ctx, tx, originalJob, report, now)
}
