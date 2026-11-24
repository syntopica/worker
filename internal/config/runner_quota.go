package config

// RunnerQuota is where a runner's quota is read from.
type RunnerQuota struct {
	// Provider is the provider's name in the usage reader.
	Provider string
	// Windows holds the labelled rate windows to read; empty means the
	// provider's unlabelled ones.
	Windows []string
}
