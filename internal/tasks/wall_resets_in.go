package tasks

import (
	"regexp"
	"strconv"
)

const (
	// afterWall is how far past the wall sentence the duration may start:
	// elsewhere in the output it may be the model's prose.
	afterWall = 200
	// maxWallS is the longest reset trusted as a duration: 14 days.
	maxWallS = 14 * 24 * 3600.0
)

var resetsIn = regexp.MustCompile(`(?i)resets in (?:(\d+)h)?(?:(\d+)m)?(?:(\d+)s)?`)

// wallResetsIn is the seconds until a quota wall resets, read right after
// the runner's wall sentence ("... Resets in 167h30m25s.").
func wallResetsIn(wall *regexp.Regexp, output string) (seconds float64, known bool) {
	found := wall.FindStringIndex(output)
	if found == nil {
		return 0, false
	}

	parts := resetsIn.FindStringSubmatch(output[found[1]:min(len(output), found[1]+afterWall)])
	if parts == nil {
		return 0, false
	}

	for index, unit := range []float64{3600, 60, 1} {
		count, _ := strconv.ParseFloat(parts[index+1], 64) // an absent part counts zero
		seconds += count * unit
	}

	return seconds, seconds > 0 && seconds <= maxWallS
}
