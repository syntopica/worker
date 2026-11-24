package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/store"
)

const (
	minJudgeAnswers = 2
	judgeAttempts   = 2
)

func submitJudge(ctx context.Context, tx store.Conn, cfg *config.Config, group string, now float64) error {
	var id string
	_, err := queryRow(ctx, tx, "shadow_first_member", params{keyGroup: group}, &id)
	if err != nil {
		return err
	}
	first, err := loadJob(ctx, tx, id)
	if err != nil {
		return err
	}
	shadow := cfg.Queues[first.Queue].Shadow
	rows, err := queryMaps(ctx, tx, "shadow_answers", params{keyGroup: group})
	if err != nil {
		return err
	}
	member := &shadowMember{group: group, producer: producerJudge, kind: kindTask, pin: pinJudge, maxAttempts: 1}
	if len(rows) < minJudgeAnswers || shadow == nil {
		return cancelJudgeMarker(ctx, tx, first, member, now)
	}
	input, err := loadJobInput(ctx, tx, text(rows[0][keyID]))
	if err != nil {
		return err
	}
	if input == nil {
		return cancelJudgeMarker(ctx, tx, first, member, now)
	}
	task, err := judgeTaskInput(group, input, rows)
	if err != nil {
		return err
	}
	member.model, member.maxAttempts = shadow.Judge, judgeAttempts
	_, err = insertShadowJob(ctx, tx, first, task, member, now)
	return err
}

func cancelJudgeMarker(ctx context.Context, tx store.Conn, job *jobRow, member *shadowMember, now float64) error {
	id, err := insertShadowJob(ctx, tx, job, map[string]any{}, member, now)
	if err != nil {
		return err
	}
	return run(ctx, tx, "cancel_judge_marker", params{keyNow: now, keyID: id})
}
