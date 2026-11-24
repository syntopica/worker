package config

// ShadowPolicy is how a queue samples its answers for judged comparison.
type ShadowPolicy struct {
	// Rate is the share of succeeded inference jobs shadowed.
	Rate float64
	// Targets are "ollama:<model>", "openrouter:<model>" or
	// "runner:<profile>".
	Targets []string
	// Judge is the task profile that scores the answers.
	Judge string
	// MaxPending is the unfinished shadow members above which no new group
	// opens.
	MaxPending int
}
