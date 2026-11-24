package config

type rawNode struct {
	Trust              *string  `json:"trust"`
	IdleThresholdS     *float64 `json:"idle_threshold_s"`
	MemoryBudgetGB     *float64 `json:"memory_budget_gb"`
	OllamaURL          *string  `json:"ollama_url"`
	OllamaLaunchdLabel *string  `json:"ollama_launchd_label"`
	MinFreePct         *float64 `json:"min_free_pct"`
	OpenRouterKeyFile  string   `json:"openrouter_key_file"`
	RemoteSlots        int      `json:"remote_slots"`
	TaskSlots          int      `json:"task_slots"`
	MaxLoad            *float64 `json:"max_load"`
	CoordinatorURL     string   `json:"coordinator_url"`
	LocalIdleOnly      bool     `json:"local_idle_only"`
}

func (n *rawNode) complete() bool {
	return n.Trust != nil && n.MemoryBudgetGB != nil && n.OllamaURL != nil && n.OllamaLaunchdLabel != nil
}
