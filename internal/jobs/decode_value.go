package jobs

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"unicode/utf8"
)

var errNotOneJSONValue = errors.New("jobs: not one UTF-8 JSON value")

// decodeValue decodes one JSON value with its numbers kept as written, so an
// integer survives storage and hashing unchanged.
func decodeValue(raw []byte) (any, error) {
	if !utf8.Valid(raw) {
		return nil, errNotOneJSONValue
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, fmt.Errorf("jobs: decode JSON: %w", err)
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return nil, errNotOneJSONValue
	}

	return value, nil
}
