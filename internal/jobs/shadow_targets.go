package jobs

import (
	"strings"

	"github.com/syntopica/worker/internal/config"
)

func shadowTargets(cfg *config.Config, privacy string, targets []string, executor map[string]any) []string {
	var allowed []string
	for _, target := range targets {
		if targetAllowed(cfg, privacy, target) && !targetAnswered(cfg, target, executor) {
			allowed = append(allowed, target)
		}
	}
	return allowed
}

func targetAnswered(cfg *config.Config, target string, executor map[string]any) bool {
	kind, name, _ := strings.Cut(target, ":")
	if kind == executorRunner {
		profile, found := cfg.Profiles[name]
		return found && profile.Runner == executor[keyProvider]
	}
	return kind == executor[keyProvider] && name == executor[keyModel]
}
