package jobs

import "strings"

// inferenceTaskInput renders a chat as the prompt understood by task runners.
func inferenceTaskInput(input map[string]any) (map[string]any, error) {
	if err := checkInferenceInput(input); err != nil {
		return nil, errInvalidInput
	}
	messages, _ := input[keyMessages].([]any)
	prompt := make([]string, 0, len(messages))
	for _, value := range messages {
		message := object(value)
		prompt = append(prompt, "["+text(message["role"])+"]\n"+text(message["content"]))
	}
	task := map[string]any{keyPrompt: strings.Join(prompt, "\n\n")}
	if schema := input[keySchema]; schema != nil {
		task[keyOutputSchema] = schema
	}
	return task, nil
}
