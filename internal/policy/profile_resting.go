package policy

import "github.com/syntopica/worker/internal/config"

// ProfileResting reports whether a task profile rests after a quota wall:
// its whole runner does, or the model it pins.
func ProfileResting(profile *config.TaskProfile, cooling map[string]bool) bool {
	return cooling[profile.Runner] || cooling[config.CooldownKey(profile.Runner, profile.Model)]
}
