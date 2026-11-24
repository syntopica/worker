package jobs

// checkInferenceInput refuses with bad_input unless the node can run the
// input as given: messages is a non-empty list of {role, content} strings,
// options and schema are objects when present, and format is "json".
func checkInferenceInput(input map[string]any) error {
	messages, _ := input[keyMessages].([]any)
	if len(messages) == 0 {
		return bad(codeBadInput)
	}
	for _, message := range messages {
		fields := object(message)
		_, hasRole := fields["role"].(string)
		_, hasContent := fields["content"].(string)
		if !hasRole || !hasContent {
			return bad(codeBadInput)
		}
	}
	for _, name := range []string{"options", keySchema} {
		value, present := input[name]
		if _, isObject := value.(map[string]any); present && !isObject {
			return bad(codeBadInput)
		}
	}
	if format, present := input["format"]; present && format != "json" {
		return bad(codeBadInput)
	}

	return nil
}
