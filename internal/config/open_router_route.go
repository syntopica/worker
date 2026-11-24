package config

import "time"

// OpenRouterRoute is how a queue escalates inference to OpenRouter's free
// endpoints (spec 8, rung 3).
type OpenRouterRoute struct {
	// Models maps a local model to the OpenRouter model that may stand in
	// for it.
	Models map[string]string
	// After is how long a job waits since it was created before it escalates.
	After time.Duration
	// ZDR sends a non-public job only to zero-data-retention endpoints;
	// turning it off is an explicit owner decision for the queue.
	ZDR bool
	// Fallbacks are further free models OpenRouter tries, in order, when the
	// mapped one errors or is rate limited.
	Fallbacks []string
	// DailyCap bounds the queue's OpenRouter attempts per UTC day; nil
	// leaves it unbounded.
	DailyCap *int
	// DailyKeyCap bounds the key instead: the queue stops once every queue's
	// OpenRouter attempts that UTC day reach it.
	DailyKeyCap *int
}
