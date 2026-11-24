package remote

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/contract"
	"github.com/syntopica/worker/internal/policy"
)

const (
	provider = "openrouter"

	noHeadroomRest  = 900 * time.Second
	idleRest        = 30 * time.Second
	minGap          = 3100 * time.Millisecond // 20 requests per minute on free endpoints
	rateLimitedRest = 300 * time.Second
	errorRest       = 60 * time.Second
)

// errUnknownNode is a loop whose node the configuration does not name.
var errUnknownNode = errors.New("unknown_node")

// Loop is the OpenRouter loop: one escalated inference job at a time. Every
// field is required.
type Loop struct {
	// Name is the node's name in the configuration.
	Name string
	// Config returns the current configuration; it is read every iteration.
	Config func() *config.Config
	// Link is the coordinator.
	Link Coordinator
	// Remote is the endpoint the jobs are sent to.
	Remote Provider
	// Now is the clock.
	Now func() time.Time
	// Sleep waits, or returns early when the context ends.
	Sleep func(ctx context.Context, wait time.Duration)
	// Beat is how often a running attempt heartbeats.
	Beat time.Duration
	// Log receives the loop's log lines: fixed words only, never content.
	Log io.Writer
}

// Run loops until the context ends and returns its error. A failing
// iteration logs its fixed error word, rests, and never stops the loop.
func (l *Loop) Run(ctx context.Context) error {
	for ctx.Err() == nil {
		rest, err := l.step(ctx)
		if err != nil {
			l.logf("worker: remote iteration failed: %s\n", err)

			rest = errorRest
		}

		l.Sleep(ctx, rest)
	}

	return fmt.Errorf("remote loop stopped: %w", ctx.Err())
}

// step is one iteration: check headroom, lease, run. It returns how long to
// rest before the next one.
//
// Unknown headroom is no headroom (spec 8): the loop waits rather than
// spending a request it cannot account for. The loop checks the lease's
// privacy class itself, which catches a coordinator and node that disagree.
func (l *Loop) step(ctx context.Context) (time.Duration, error) {
	cfg := l.Config()

	node, known := cfg.Nodes[l.Name]
	if !known {
		return 0, errUnknownNode
	}

	if left, counted := l.Remote.Headroom(ctx); !counted || left <= 0 {
		return noHeadroomRest, nil
	}

	lease := l.Link.LeaseRemote(ctx)
	if lease == nil {
		return idleRest, nil
	}

	run := &attempt{
		loop:     l,
		lease:    lease,
		executor: contract.Executor{Node: l.Name, Provider: provider, Model: lease.Model},
		started:  l.Now(),
	}

	how, why := failed, privacyRefused
	if policy.PrivacyAllows(cfg, lease.Privacy, provider, node.Trust) {
		how, why = run.run(ctx, cfg)
	} else {
		run.complete(ctx, how, why, nil)
	}

	if how != succeeded {
		// Outcome and code are fixed allowlisted words; the job id is opaque.
		l.logf("worker: remote %s: %s job=%s\n", how, why, lease.JobID)
	}

	if why == rateLimited {
		return rateLimitedRest, nil
	}

	return minGap, nil
}

// logf writes one log line; a failed write is not worth stopping for.
func (l *Loop) logf(format string, args ...any) {
	_, _ = fmt.Fprintf(l.Log, format, args...)
}
