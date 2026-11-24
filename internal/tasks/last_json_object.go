package tasks

import (
	"encoding/json"
	"strings"
)

// lastJSONObject is the last balanced JSON object in text, or empty: it
// scans from each opening brace and keeps the last object that parses.
func lastJSONObject(text string) string {
	last := ""
	index := strings.IndexByte(text, '{')

	for index != -1 {
		decoder := json.NewDecoder(strings.NewReader(text[index:]))

		var object json.RawMessage
		if decoder.Decode(&object) != nil {
			index = next(text, index+1)

			continue
		}

		last = string(object)
		index = next(text, index+int(decoder.InputOffset()))
	}

	return last
}

// next is the index of the first opening brace at or after from, or -1.
func next(text string, from int) int {
	if found := strings.IndexByte(text[from:], '{'); found != -1 {
		return from + found
	}

	return -1
}
