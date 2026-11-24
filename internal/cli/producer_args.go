package cli

import (
	"math"
	"slices"
	"time"
)

type producerArgs struct {
	principal, queue string
	values           []string
	wait, poll       time.Duration
}

func (a *producerArgs) valid(name string, wait, poll float64) bool {
	const maxSeconds = float64(math.MaxInt64) / float64(time.Second)
	switch {
	case a.principal == "", name == jobsCommand && a.queue == "":
		return false
	case name == rateCommand:
		return slices.Contains([]string{"good", "edited", "discarded"}, a.values[2])
	case name == runCommand:
		return wait > -maxSeconds && wait < maxSeconds && poll >= 0 && poll < maxSeconds
	default:
		return true
	}
}
