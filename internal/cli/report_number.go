package cli

import "encoding/json"

func reportNumber(number json.Number) string {
	if number == "" {
		return "None"
	}
	return string(number)
}
