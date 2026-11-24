package config

const (
	defaultIdleThresholdS = 300
	defaultMinFreePct     = 15.0
)

// parseNode builds one node policy, applying the defaults.
func parseNode(name string, raw *rawNode) NodePolicy {
	return NodePolicy{
		Name:               name,
		Trust:              raw.Trust,
		IdleThreshold:      seconds(orDefault(raw.IdleThresholdS, defaultIdleThresholdS)),
		MemoryBudgetGB:     raw.MemoryBudgetGB,
		OllamaURL:          raw.OllamaURL,
		OllamaLaunchdLabel: raw.OllamaLaunchdLabel,
		MinFreePct:         orDefault(raw.MinFreePct, defaultMinFreePct),
		MaxLoad:            raw.MaxLoad,
		LocalIdleOnly:      raw.LocalIdleOnly,
		OpenRouterKeyFile:  raw.OpenRouterKeyFile,
		RemoteSlots:        max(1, raw.RemoteSlots),
		TaskSlots:          max(1, raw.TaskSlots),
		CoordinatorURL:     raw.CoordinatorURL,
	}
}
