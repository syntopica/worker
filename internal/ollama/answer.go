package ollama

import "encoding/json"

// Answer is a finished chat call: the joined content and the counts of the
// final chunk.
type Answer struct {
	// Content is the content of every chunk, joined.
	Content string
	// TokensIn is the prompt token count, 0 when the server gave none.
	TokensIn int
	// TokensOut is the generated token count, 0 when the server gave none.
	TokensOut int
}

// Output is the contract's output: the text, and the text as JSON when it is
// a JSON document, else null.
func (a Answer) Output() map[string]any {
	var parsed json.RawMessage
	if json.Valid([]byte(a.Content)) {
		parsed = json.RawMessage(a.Content)
	}

	return map[string]any{"text": a.Content, "json": parsed}
}
