package cli

import (
	"encoding/json"
	"fmt"
	"io"
)

func writeJSON(out io.Writer, value any, pretty bool) error {
	enc := json.NewEncoder(out)
	enc.SetEscapeHTML(false)
	if pretty {
		enc.SetIndent("", "  ")
	}
	if err := enc.Encode(value); err != nil {
		return fmt.Errorf("write JSON: %w", err)
	}
	return nil
}
