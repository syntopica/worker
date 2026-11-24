package tasks

import (
	"strings"
	"testing"
)

func TestAWallNamesItsResetOnlyRightAfterItsSentence(t *testing.T) {
	t.Parallel()

	const wall = "error: Individual quota reached. Please upgrade your subscription to increase your limits."

	cases := map[string]float64{
		wall + " Resets in 167h30m25s.":                                      167*3600 + 30*60 + 25,
		wall + " Resets in 45m.":                                             45 * 60,
		"The model wrote: Resets in 300h. " + strings.Repeat("x", 50) + wall: 0,
		"Resets in 45m":                                                      0,
		wall + " Resets in 9999h.":                                           0,
		wall + " Resets in 0h.":                                              0,
		wall + " It resets one day":                                          0,
	}

	for output, want := range cases {
		seconds, known := wallResetsIn(agy().wall, output)
		if known != (want > 0) || (known && seconds != want) {
			t.Errorf("%.60q: %v %v", output, seconds, known)
		}
	}
}
