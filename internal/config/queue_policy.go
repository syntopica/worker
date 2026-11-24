package config

// QueuePolicy is how one queue is scheduled and routed.
type QueuePolicy struct {
	// Name is the queue's name.
	Name string
	// OpenRouter lets the queue's inference jobs escalate to free remote
	// endpoints; nil means they never do.
	OpenRouter *OpenRouterRoute
}
