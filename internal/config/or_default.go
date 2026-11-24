package config

// orDefault is the configured value, or fallback when it is absent.
func orDefault[T any](value *T, fallback T) T {
	if value == nil {
		return fallback
	}

	return *value
}
