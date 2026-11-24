package jsonanswer

import (
	"encoding/json"
	"strings"
)

// Parse returns the answer as a JSON document, or nil when it is not one.
//
// Some free models and runners wrap JSON in a ```json fence even when asked
// for JSON only; a single surrounding fence is removed before parsing.
func Parse(text string) json.RawMessage {
	body := strings.TrimSpace(text)
	if strings.HasPrefix(body, "```") && strings.HasSuffix(body, "```") {
		body = strings.TrimSpace(strings.TrimPrefix(strings.Trim(body, "`"), "json"))
	}

	if !json.Valid([]byte(body)) {
		return nil
	}

	return json.RawMessage(body)
}
