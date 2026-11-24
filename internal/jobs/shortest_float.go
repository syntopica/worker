package jobs

import (
	"math"
	"strconv"
	"strings"
)

const (
	fixedFrom  = -4
	fixedBelow = 16
)

// shortestFloat writes a float as "1.0", "0.0001", "1e+16" or "1.5e-05":
// fixed notation for exponents from -4 up to 15, scientific outside.
func shortestFloat(number float64) string {
	if math.IsInf(number, 0) {
		return strings.TrimPrefix(strconv.FormatFloat(number, 'f', -1, 64), "+") + "inity"
	}
	scientific := strconv.FormatFloat(number, 'e', -1, 64)
	exponent, _ := strconv.Atoi(scientific[strings.IndexByte(scientific, 'e')+1:])
	if exponent < fixedFrom || exponent >= fixedBelow {
		return scientific
	}
	fixed := strconv.FormatFloat(number, 'f', -1, 64)
	if !strings.Contains(fixed, ".") {
		fixed += ".0"
	}

	return fixed
}
