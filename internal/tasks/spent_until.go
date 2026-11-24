package tasks

import (
	"strings"
	"time"
)

// spentPercent is the usage from which a rate window counts as spent.
const spentPercent = 95

// spentUntil is when (Unix seconds) a usage report says a provider's quota
// comes back: the latest reset among its spent windows.
//
// With windows only the labelled extra windows whose id or title contains
// one of them are read (one provider can meter model families separately);
// otherwise the unlabelled primary, secondary and tertiary windows. A window
// whose usage is unknown is no evidence either way.
func spentUntil(usage map[string]any, windows []string, now time.Time) (until float64, spent bool) {
	var read []any

	if len(windows) == 0 {
		read = []any{usage["primary"], usage["secondary"], usage["tertiary"]}
	}

	extras, _ := usage["extraRateWindows"].([]any)
	for _, entry := range extras {
		extra, _ := entry.(map[string]any)
		id, _ := extra["id"].(string)
		title, _ := extra["title"].(string)
		label := strings.ToLower(id + " " + title)

		for _, window := range windows {
			if strings.Contains(label, strings.ToLower(window)) {
				read = append(read, extra["window"])

				break
			}
		}
	}

	for _, entry := range read {
		window, _ := entry.(map[string]any)
		used, _ := window["usedPercent"].(float64)
		resets, _ := window["resetsAt"].(string)

		end, err := time.Parse(time.RFC3339, resets)
		if window["usageKnown"] == false || used < spentPercent || err != nil || !end.After(now) {
			continue
		}

		until, spent = max(until, float64(end.UnixMilli())/float64(time.Second/time.Millisecond)), true
	}

	return until, spent
}
