package jobs

import (
	"slices"
	"strings"

	"github.com/syntopica/worker/internal/config"
)

func targetAllowed(cfg *config.Config, privacy, target string) bool {
	kind, name, _ := strings.Cut(target, ":")
	if !slices.Contains(cfg.Privacy[privacy], kind) {
		return false
	}
	return kind != executorRunner || slices.Contains(cfg.Profiles[name].Privacy, privacy)
}
