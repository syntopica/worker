package jobs

import (
	"encoding/json"
	"math"
)

// sqlValue is a reported JSON scalar as a column value; anything that is not
// a scalar is stored as NULL.
func sqlValue(value any) any {
	switch typed := value.(type) {
	case json.Number:
		if whole, err := typed.Int64(); err == nil {
			return whole
		}
		if number, err := typed.Float64(); err == nil {
			return number
		}

		return nil
	case int:
		return int64(typed)
	case int64:
		return typed
	case float64:
		if !math.IsNaN(typed) && !math.IsInf(typed, 0) {
			return typed
		}
		return nil
	default:
		return nil
	}
}
