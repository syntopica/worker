package cli

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
)

func renderStatus(raw json.RawMessage) (string, error) {
	var report struct {
		Queues map[string]struct {
			States map[string]int64 `json:"states"`
			Done   int64            `json:"done_1h"`
			Wasted float64          `json:"wasted_1h_s"`
			Oldest json.Number      `json:"oldest_queued_s"`
		} `json:"queues"`
		Cooldowns map[string]float64 `json:"cooldowns"`
		Remote    *struct {
			Spent  int64       `json:"spent_today"`
			KeyCap json.Number `json:"key_cap"`
			Capped []string    `json:"capped"`
		} `json:"openrouter"`
	}
	if err := json.Unmarshal(raw, &report); err != nil {
		return "", fmt.Errorf("read queue status: %w", err)
	}

	var out strings.Builder
	for _, name := range slices.Sorted(maps.Keys(report.Queues)) {
		queue := report.Queues[name]
		fmt.Fprintf(&out, "%-24s %s done_1h=%d wasted_1h=%.0fs oldest=%s\n",
			name, renderQueueStates(queue.States), queue.Done, queue.Wasted, reportNumber(queue.Oldest))
	}
	for _, runner := range slices.Sorted(maps.Keys(report.Cooldowns)) {
		fmt.Fprintf(&out, "cooldown %-15s %.0fs left\n", runner, report.Cooldowns[runner])
	}
	if remote := report.Remote; remote != nil {
		capacity := string(remote.KeyCap)
		if capacity == "" || capacity == "0" {
			capacity = "-"
		}
		capped := strings.Join(remote.Capped, ",")
		if capped == "" {
			capped = "none"
		}
		fmt.Fprintf(&out, "openrouter today %d/%s capped=%s\n", remote.Spent, capacity, capped)
	}
	return out.String(), nil
}

func renderQueueStates(states map[string]int64) string {
	entries := make([]string, 0, len(states))
	for _, state := range slices.Sorted(maps.Keys(states)) {
		entries = append(entries, strconv.Quote(state)+": "+strconv.FormatInt(states[state], 10))
	}
	return "{" + strings.Join(entries, ", ") + "}"
}
