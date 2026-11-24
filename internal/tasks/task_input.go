package tasks

import "encoding/json"

// taskInput is a task job's input (spec 8).
type taskInput struct {
	// Prompt is what the runner is asked.
	Prompt string `json:"prompt"`
	// Inputs is the manifest: paths under the profile's input root.
	Inputs []string `json:"inputs"`
	// OutputSchema is the JSON schema the answer must match, if any.
	OutputSchema json.RawMessage `json:"output_schema"`
}

// schema is the output schema as JSON text, or empty when there is none.
func (t *taskInput) schema() string {
	if string(t.OutputSchema) == "null" {
		return ""
	}

	return string(t.OutputSchema)
}
