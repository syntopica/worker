package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/syntopica/worker/internal/config"
)

// RequestBody builds an /api/chat body with the model's pinned options (spec
// 5): num_ctx and keep_alive always come from the pin, never the job.
//
// The answer is streamed so the node knows whether a cancel lands in prefill
// or generation (the 1c cancel-to-quiet measurement).
func RequestBody(pin *config.ModelPin, input json.RawMessage) (map[string]any, error) {
	var job struct {
		Messages json.RawMessage            `json:"messages"`
		Options  map[string]json.RawMessage `json:"options"`
		Schema   json.RawMessage            `json:"schema"`
		Format   json.RawMessage            `json:"format"`
	}

	if err := json.Unmarshal(input, &job); err != nil {
		return nil, fmt.Errorf("decode job input: %w", err)
	}

	options := map[string]any{}

	for key, value := range job.Options {
		if slices.Contains(samplingKeys(), key) {
			options[key] = value
		}
	}

	options["num_ctx"] = pin.NumCtx

	messages := job.Messages
	if empty(messages) {
		messages = json.RawMessage("[]")
	}

	body := map[string]any{
		"model":      pin.Name,
		"messages":   messages,
		"stream":     true,
		"think":      false,
		"keep_alive": pin.KeepAlive,
		"options":    options,
	}

	switch {
	case !empty(job.Schema):
		body["format"] = job.Schema
	case string(bytes.TrimSpace(job.Format)) == `"json"`:
		body["format"] = "json"
	}

	return body, nil
}

// samplingKeys are the only options a job may set.
func samplingKeys() []string {
	return []string{"temperature", "top_p", "top_k", "seed", "num_predict", "repeat_penalty"}
}

// empty reports whether a JSON value is absent or one a job leaves unset:
// null, false, zero, or an empty string, list or object.
func empty(raw json.RawMessage) bool {
	return slices.Contains([]string{"", "null", "false", "0", `""`, "[]", "{}"}, string(bytes.TrimSpace(raw)))
}
