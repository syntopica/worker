package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/syntopica/worker/internal/contract"
)

const unknown = "unknown"

// decode turns a non-2xx answer into an *Error and reads a 2xx body into out
// when out is set.
func decode(response *http.Response, out any) error {
	raw, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("read answer: %w", err)
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return &Error{Status: response.StatusCode, Code: errorCode(raw)}
	}

	if out == nil || len(raw) == 0 {
		return nil
	}

	if err = json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("decode answer: %w", err)
	}

	return nil
}

// errorCode is the "error" field of an object body, or "unknown".
func errorCode(raw []byte) string {
	var refusal contract.Refusal
	if err := json.Unmarshal(raw, &refusal); err != nil || refusal.Error == "" {
		return unknown
	}

	return refusal.Error
}
