package jobs

// object is the value as a JSON object; anything else reads as an empty one.
func object(value any) map[string]any {
	fields, _ := value.(map[string]any)

	return fields
}
