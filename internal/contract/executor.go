package contract

// Executor names what ran an attempt.
type Executor struct {
	// Node is the machine that ran it.
	Node string `json:"node"`
	// Provider is the backend, such as "ollama" or "openrouter".
	Provider string `json:"provider"`
	// Model is the model that answered.
	Model string `json:"model"`
}
