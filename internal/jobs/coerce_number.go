package jobs

import (
	"encoding/json"
	"math"
	"strconv"
	"strings"
)

// coerceNumber reads a request field as a number and refuses anything else
// with out_of_range. A whole number drops the fraction of a JSON float and
// accepts a string only when it is an integer literal.
func coerceNumber(value any, whole bool) (float64, error) {
	var literal string
	integerOnly := false
	switch typed := value.(type) {
	case json.Number:
		literal = typed.String()
	case string:
		literal, integerOnly = strings.TrimSpace(typed), whole
	default:
		return 0, bad(codeOutOfRange)
	}
	if integerOnly || whole && !strings.ContainsAny(literal, ".eE") {
		number, err := strconv.ParseInt(literal, 10, 64)
		if err != nil {
			return 0, bad(codeOutOfRange)
		}
		return float64(number), nil
	}
	return coerceFloat(literal, whole)
}

func coerceFloat(literal string, whole bool) (float64, error) {
	number, err := strconv.ParseFloat(literal, 64)
	if err != nil || math.IsInf(number, 0) || math.IsNaN(number) {
		return 0, bad(codeOutOfRange)
	}
	if whole {
		number = math.Trunc(number)
	}

	return number, nil
}
