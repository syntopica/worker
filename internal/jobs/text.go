package jobs

import (
	"encoding/json"
	"fmt"
)

// text is the value as a string; null and absent read as empty.
func text(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case json.Number:
		return typed.String()
	case nil:
		return ""
	default:
		return fmt.Sprint(typed)
	}
}
