package config

import (
	"errors"
	"fmt"
	"time"
)

const defaultCooldownS = 3600

var errModelWindows = errors.New("model_windows must map windows to model fragments")

// fillRunners reads the per-runner settings: cooldown, quota source, the
// model fragments each rate window meters and the concurrency cap.
func fillRunners(config *Config, runners map[string]rawRunner) error {
	config.RunnerCooldown = make(map[string]time.Duration, len(runners))
	config.RunnerQuota = map[string]RunnerQuota{}
	config.RunnerModelWindows = map[string]map[string][]string{}
	config.RunnerMaxConcurrent = map[string]int{}

	for name := range runners {
		runner := runners[name]
		config.RunnerCooldown[name] = seconds(orDefault(runner.CooldownS, defaultCooldownS))

		if runner.QuotaProvider != "" {
			config.RunnerQuota[name] = RunnerQuota{Provider: runner.QuotaProvider, Windows: runner.QuotaWindows}
		}

		if runner.MaxConcurrent != nil {
			config.RunnerMaxConcurrent[name] = *runner.MaxConcurrent
		}

		if runner.ModelWindows == nil {
			continue
		}

		for _, fragments := range runner.ModelWindows {
			if len(fragments) == 0 {
				return fmt.Errorf("runner %s: %w", name, errModelWindows)
			}
		}

		config.RunnerModelWindows[name] = runner.ModelWindows
	}

	return nil
}
