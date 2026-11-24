package jobs

import (
	"context"
	"slices"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/store"
)

// jobCoolingUntil returns the first of the available profiles' cooldown ends
// when all of them are resting.
func jobCoolingUntil(ctx context.Context, tx store.Conn, cfg *config.Config, job *jobRow, now float64) (*float64, error) {
	var until *float64
	if job.Kind != kindTask || job.State != stateQueued {
		return until, nil
	}
	names := slices.Concat([]string{job.Model}, cfg.Queues[job.Queue].Fallbacks[job.Model])
	for _, name := range names {
		profile, found := cfg.Profiles[name]
		if !found {
			continue
		}
		var end *float64
		_, err := queryRow(ctx, tx, "cooldown_end", params{
			keyRunner: profile.Runner, "model_key": config.CooldownKey(profile.Runner, profile.Model), keyNow: now,
		}, &end)
		if err != nil || end == nil {
			return end, err
		}
		if until == nil || *end < *until {
			until = end
		}
	}
	return until, nil
}
