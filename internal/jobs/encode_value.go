package jobs

import (
	"encoding/json"
	"fmt"
)

// encodeValue is the stored form of a JSON value.
func encodeValue(value any) (string, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("jobs: encode JSON: %w", err)
	}

	return string(raw), nil
}
