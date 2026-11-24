package tasks

import (
	"context"
	"strings"
	"time"

	"github.com/syntopica/worker/internal/config"
)

// probeWalls is until when (Unix seconds) each configured runner known to be
// spent now rests. A runner with model windows rests per model instead,
// under its "runner:model" key.
func probeWalls(
	ctx context.Context, cfg *config.Config, now time.Time, read func(ctx context.Context, provider string) map[string]any,
) map[string]float64 {
	walls := map[string]float64{}

	for runner, quota := range cfg.RunnerQuota {
		usage := read(ctx, quota.Provider)
		if usage == nil {
			continue
		}

		if windows := cfg.RunnerModelWindows[runner]; len(windows) > 0 {
			modelWalls(walls, runner, usage, windows, cfg.Profiles, now)

			continue
		}

		if until, spent := spentUntil(usage, quota.Windows, now); spent {
			walls[runner] = until
		}
	}

	return walls
}

// modelWalls adds a rest for each profile model a spent window meters.
//
// windows maps a window label to the model-name fragments it meters, so a
// spent window rests the runner's models of that family and leaves the
// others it serves from a separate allowance free to run.
func modelWalls(
	walls map[string]float64, runner string, usage map[string]any, windows map[string][]string,
	profiles map[string]config.TaskProfile, now time.Time,
) {
	for window, fragments := range windows {
		until, spent := spentUntil(usage, []string{window}, now)
		if !spent {
			continue
		}

		for name := range profiles {
			model := strings.ToLower(profiles[name].Model)
			if profiles[name].Runner != runner || model == "" || !meters(fragments, model) {
				continue
			}

			key := config.CooldownKey(runner, profiles[name].Model)
			walls[key] = max(walls[key], until)
		}
	}
}

// meters reports whether a lowercase model name holds one of the fragments.
func meters(fragments []string, model string) bool {
	for _, fragment := range fragments {
		if strings.Contains(model, strings.ToLower(fragment)) {
			return true
		}
	}

	return false
}
