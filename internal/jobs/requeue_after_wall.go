package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/store"
)

func requeueAfterWall(
	ctx context.Context, tx store.Conn, cfg *config.Config, job *jobRow, report *CompletionReport, now float64,
) error {
	var selected *string
	if _, err := queryRow(ctx, tx, "attempt_profile", params{keyAttempt: job.LeaseAttempt}, &selected); err != nil {
		return err
	}
	name := job.Model
	if selected != nil {
		name = *selected
	}
	profile := cfg.Profiles[name]
	runner, model := profile.Runner, profile.Model
	if job.Kind == kindInference {
		runner, model = text(report.Executor[keyProvider]), text(report.Executor[keyModel])
	}
	seconds := hourSeconds
	if cooldown, ok := cfg.RunnerCooldown[runner]; ok {
		seconds = cooldown.Seconds()
	}
	until := now + max(seconds, trustedWallReset(report.Usage[keyWallReset]))
	if runner != "" {
		if err := run(ctx, tx, "rest_runner", params{keyRunner: config.CooldownKey(runner, model), "until": until}); err != nil {
			return err
		}
	}
	due := until
	if job.Kind == kindInference || name != job.Model || len(cfg.Queues[job.Queue].Fallbacks[job.Model]) > 0 {
		due = now
	}
	return requeueJob(ctx, tx, job.ID, codeQuotaWall, now, due)
}
