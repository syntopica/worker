package node

import (
	"fmt"
	"io"
	"slices"
	"time"
)

// residencyWindow is how far back loads and unloads are counted.
const residencyWindow = time.Hour

// trackResidency records residency changes since the last sample and returns
// the last hour's counts, for the reload criterion of 1c.
//
// A change is counted whatever caused it (a lease, a pressure unload, a
// keep-alive expiry or another application), because the acceptance
// criterion is about reloads, not about who caused them. Each change is also
// logged with its time.
func trackResidency(m *memory, resident []string, now time.Time, log io.Writer) (loads, unloads int) {
	current := map[string]bool{}
	for _, model := range resident {
		current[model] = true
	}

	if m.lastResident != nil {
		recordChanges(m, current, m.lastResident, true, now, log)
		recordChanges(m, m.lastResident, current, false, now, log)
	}

	m.lastResident = current
	m.residency = slices.DeleteFunc(m.residency, func(event residencyEvent) bool {
		return now.Sub(event.at) >= residencyWindow
	})

	for _, event := range m.residency {
		if event.load {
			loads++
		}
	}

	return loads, len(m.residency) - loads
}

// recordChanges logs and records the models in one sample and not the other.
func recordChanges(m *memory, in, notIn map[string]bool, load bool, now time.Time, log io.Writer) {
	kind := "unload"
	if load {
		kind = "load"
	}

	var changed []string

	for model := range in {
		if !notIn[model] {
			changed = append(changed, model)
		}
	}

	slices.Sort(changed)

	for _, model := range changed {
		_, _ = fmt.Fprintf(log, "worker: residency %s model=%s at=%s\n", kind, model, now.UTC().Format(stampLayout))

		m.residency = append(m.residency, residencyEvent{at: now, load: load})
	}
}
