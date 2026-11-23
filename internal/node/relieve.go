package node

import (
	"context"

	"github.com/syntopica/worker/internal/ollama"
)

// relieve unloads this node's own models when the host needs the memory
// back. Only owned models qualify: never one resident at startup or loaded
// by another application.
func relieve(ctx context.Context, backend ollama.Backend, m *memory, resident []string, onAC *bool, ownerActive bool) {
	// Sustained pressure: once per pressure episode.
	if m.pressureStreak >= pressureSamples && !m.unloaded {
		m.unloaded = unloadOwned(ctx, backend, m, resident, nil)
	}

	// Battery lasts, unlike a short user_active gap, so the warm model buys
	// nothing; an unknown power state unloads nothing.
	if onAC != nil && *onAC {
		clear(m.batteryTried)
	}

	if onAC != nil && !*onAC {
		unloadOwned(ctx, backend, m, resident, m.batteryTried)
	}

	// On a local-idle-only node the warm model buys nothing while the owner
	// is active, so it goes as soon as they are.
	if !ownerActive {
		clear(m.activeTried)

		return
	}

	unloadOwned(ctx, backend, m, resident, m.activeTried)
}

// unloadOwned unloads the resident models the node owns and reports whether
// there were any. With tried, each model gets one attempt per episode: a
// failed unload is not retried every step and keeps its ownership, and a
// failed drain's model is left to retryFailedDrain.
func unloadOwned(ctx context.Context, backend ollama.Backend, m *memory, resident []string, tried map[string]bool) bool {
	var mine []string

	for _, model := range resident {
		if m.owned[model] && (tried == nil || (model != m.failedModel && !tried[model])) {
			mine = append(mine, model)
		}
	}

	for _, model := range mine {
		if tried != nil {
			tried[model] = true
		}

		if backend.Unload(ctx, model) {
			delete(m.owned, model)
		}
	}

	return len(mine) > 0
}
