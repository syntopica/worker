package config

// CooldownKey is the cooldown a quota wall rests: "runner:model" when the
// profile pins a model, else the runner alone.
//
// One runner meters its models separately, so a wall rests only the model
// that hit it; a runner-wide reading rests the runner key.
func CooldownKey(runner, model string) string {
	if model == "" {
		return runner
	}

	return runner + ":" + model
}
