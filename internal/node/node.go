package node

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/host"
	"github.com/syntopica/worker/internal/ollama"
)

// restFor is the pause before the next iteration when there is nothing to do.
const restFor = 30 * time.Second

// Node is one machine's inference loop: sample, report, lease, run - one
// attempt at a time. Every field is required.
type Node struct {
	// Name is the node's name in the configuration.
	Name string
	// Config returns the current configuration; it is read every iteration.
	Config func() *config.Config
	// Link is the coordinator.
	Link Coordinator
	// Backend is the model server at a URL.
	Backend func(url string) ollama.Backend
	// HTTP sends the chat request.
	HTTP *http.Client
	// Sample reads the host state.
	Sample host.Sampler
	// Now is the clock.
	Now func() time.Time
	// Sleep waits, or returns early when the context ends.
	Sleep func(ctx context.Context, wait time.Duration)
	// Log receives the node's log lines: fixed words only, never content.
	Log io.Writer

	memory *memory
}

// Run loops until the context ends and returns its error. A failing
// iteration logs its fixed error word, rests, and never stops the loop.
func (n *Node) Run(ctx context.Context) error {
	for ctx.Err() == nil {
		rest, err := n.step(ctx)
		if err != nil {
			n.logf("worker: node iteration failed: %s\n", err)

			rest = true
		}

		if rest {
			n.Sleep(ctx, restFor)
		}
	}

	return fmt.Errorf("node stopped: %w", ctx.Err())
}

// remembered is the node's memory, started on first use.
func (n *Node) remembered() *memory {
	if n.memory == nil {
		n.memory = newMemory()
	}

	return n.memory
}

// logf writes one log line; a failed write is not worth stopping for.
func (n *Node) logf(format string, args ...any) {
	_, _ = fmt.Fprintf(n.Log, format, args...)
}
