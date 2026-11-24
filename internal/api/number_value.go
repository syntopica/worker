package api

import (
	"encoding/json"
	"math"
	"strconv"
	"strings"
)

func numberValue(input any) (float64, error) {
	switch value := input.(type) {
	case json.Number:
		return finiteNumber(string(value))
	case string:
		return finiteNumber(value)
	case bool:
		if value {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, badRequest()
	}
}

func finiteNumber(value string) (float64, error) {
	number, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil || math.IsNaN(number) || math.IsInf(number, 0) {
		return 0, badRequest()
	}
	return number, nil
}
