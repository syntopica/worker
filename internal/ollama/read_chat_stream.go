package ollama

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"strings"
)

const (
	codeBadResponse = "bad_response"
	maxChunkBytes   = 16 << 20
)

// readChatStream reads a streamed /api/chat response into one answer, or the
// error code bad_response when no done chunk arrives.
//
// The answer carries the counts of the final chunk and the content of every
// chunk joined. onChunk runs as each chunk arrives: with the first, prefill
// is over and generation has started. A chunk that is not a JSON object or
// carries "error" fails the call.
func readChatStream(body io.Reader, onChunk func()) (answer *Answer, code string) {
	var content strings.Builder

	lines := bufio.NewScanner(body)
	lines.Buffer(nil, maxChunkBytes)

	for lines.Scan() {
		raw := bytes.TrimSpace(lines.Bytes())
		if len(raw) == 0 {
			continue
		}

		onChunk()

		var chunk map[string]json.RawMessage
		if err := json.Unmarshal(raw, &chunk); err != nil || chunk == nil {
			return nil, codeBadResponse
		}

		if _, failed := chunk["error"]; failed {
			return nil, codeBadResponse
		}

		var message struct {
			Content any `json:"content"`
		}

		if json.Unmarshal(chunk["message"], &message) == nil {
			text, _ := message.Content.(string)
			content.WriteString(text)
		}

		if string(chunk["done"]) == "true" {
			return &Answer{
				Content:   content.String(),
				TokensIn:  count(chunk["prompt_eval_count"]),
				TokensOut: count(chunk["eval_count"]),
			}, ""
		}
	}

	return nil, codeBadResponse
}

// count is a token count, 0 when absent or not a number.
func count(raw json.RawMessage) int {
	var value float64
	if json.Unmarshal(raw, &value) != nil {
		return 0
	}

	return int(value)
}
