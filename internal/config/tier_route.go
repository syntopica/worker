package config

// TierRoute is what one quality tier of a queue runs on.
type TierRoute struct {
	// Models are tried before the producer's list.
	Models []string
	// Profiles maps a granted profile to its stand-in.
	Profiles map[string]string
}
