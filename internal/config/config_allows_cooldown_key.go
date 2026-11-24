package config

// AllowsCooldownKey reports whether the instance knows the runner or profile model metered by key.
func (config *Config) AllowsCooldownKey(key string) bool {
	if _, configured := config.RunnerCooldown[key]; configured {
		return true
	}
	for name := range config.Profiles {
		profile := config.Profiles[name]
		if key == profile.Runner || key == CooldownKey(profile.Runner, profile.Model) {
			return true
		}
	}

	return false
}
