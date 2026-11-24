package remote

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/syntopica/worker/internal/redact"
)

// errNoMessages is a job input without a messages list.
var errNoMessages = errors.New("job input has no messages")

// RequestBody is the chat-completions body for one escalated inference job:
// messages with credentials redacted, and the job's schema as a strict
// json_schema format.
//
// zdr restricts routing to zero-data-retention endpoints; fallbacks become
// OpenRouter's ordered models list after model.
func RequestBody(model string, input json.RawMessage, zdr bool, fallbacks []string) (map[string]any, error) {
	var job struct {
		Messages []map[string]json.RawMessage `json:"messages"`
		Options  map[string]json.RawMessage   `json:"options"`
		Schema   json.RawMessage              `json:"schema"`
	}

	if err := json.Unmarshal(input, &job); err != nil {
		return nil, fmt.Errorf("decode job input: %w", err)
	}

	if job.Messages == nil {
		return nil, errNoMessages
	}

	messages := make([]map[string]any, len(job.Messages))

	for at, message := range job.Messages {
		messages[at] = map[string]any{}
		for key, value := range message {
			messages[at][key] = value
		}

		messages[at]["content"] = redact.Credentials(contentText(message["content"]))
	}

	body := map[string]any{"model": model, "messages": messages}

	if len(fallbacks) > 0 {
		body["models"] = append([]string{model}, fallbacks...)
	}

	if temperature, set := job.Options["temperature"]; set {
		body["temperature"] = temperature
	}

	if len(job.Schema) > 0 && !bytes.Equal(bytes.TrimSpace(job.Schema), []byte("null")) {
		body["response_format"] = map[string]any{
			"type":        "json_schema",
			"json_schema": map[string]any{"name": "output", "strict": true, "schema": job.Schema},
		}
	}

	if zdr {
		body["provider"] = map[string]any{"zdr": true}
	}

	return body, nil
}

// contentText is a message's content as text: the string itself, or the raw
// JSON of anything else.
func contentText(raw json.RawMessage) string {
	var text string
	if json.Unmarshal(raw, &text) == nil {
		return text
	}

	return string(raw)
}
