package contract

// Requirements say what an inference job needs from the model that runs it.
type Requirements struct {
	// Capability names what the model must be able to do, such as "chat.json".
	Capability string `json:"capability,omitempty"`
	// Models is the ordered preference list; the first eligible one runs.
	Models []string `json:"models"`
	// MinContext is the smallest context window, in tokens, that fits the job.
	MinContext int `json:"min_context,omitempty"`
	// ExcludeProviders names the executor providers, as reported in a result's
	// executor (such as "ollama", "openrouter" or a runner name), that must
	// never run the job. Empty means none are excluded. When the exclusion
	// leaves no executor the job fails with "no_eligible_executor".
	ExcludeProviders []string `json:"exclude_providers,omitempty"`
}
