package jobs

// nullIfEmpty stores an absent name as NULL.
func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}

	return value
}
