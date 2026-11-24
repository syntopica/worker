package tasks

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/contract"
	"github.com/syntopica/worker/internal/host"
	"github.com/syntopica/worker/internal/link"
	"github.com/syntopica/worker/internal/policy"
)

const (
	// provider is the privacy executor every runner counts as.
	provider = "runner"

	rest       = 30 * time.Second
	probeEvery = 600 * time.Second
)

// errUnknownNode is a loop whose node the configuration does not name.
var errUnknownNode = errors.New("unknown_node")

// Loop is the task loop: one read-only task at a time. Every field but
// Usage is required.
type Loop struct {
	// Name is the node's name in the configuration.
	Name string
	// Config returns the current configuration; it is read every iteration.
	Config func() *config.Config
	// Link is the coordinator.
	Link Coordinator
	// Sample reads the host's state.
	Sample host.Sampler
	// Usage reads one provider's usage report, nil when it cannot be read.
	// A loop with a nil Usage does not probe quotas: with several slots only
	// the first one does.
	Usage func(ctx context.Context, provider string) map[string]any
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
//
// Every ten minutes the runners' quotas are read and any spent one is
// reported, so the coordinator rests it (and uses a queue fallback) without
// first spending a call against its wall.
func (l *Loop) Run(ctx context.Context) error {
	var nextProbe time.Time

	for ctx.Err() == nil {
		cfg := l.Config()

		if l.Usage != nil && len(cfg.RunnerQuota) > 0 && !l.Now().Before(nextProbe) {
			nextProbe = l.Now().Add(probeEvery)

			if walls := probeWalls(ctx, cfg, l.Now(), l.Usage); len(walls) > 0 {
				l.Link.ReportWalls(ctx, walls)
			}
		}

		resting, err := l.step(ctx, cfg)
		if err != nil {
			l.logf("worker: task iteration failed: %s\n", err)
		}

		if resting {
			l.Sleep(ctx, rest)
		}
	}

	return fmt.Errorf("task loop stopped: %w", ctx.Err())
}

// step is one iteration: check the host, lease a task, run it, complete it.
// It reports whether the loop should rest before the next one.
//
// Only a known battery stops admission: the runner is a CLI talking to a
// remote model, so memory pressure and an unreadable reader do not, and an
// unreadable idle time leases as an active user, which limits the lease to
// the queues that allow one.
func (l *Loop) step(ctx context.Context, cfg *config.Config) (resting bool, err error) {
	node, known := cfg.Nodes[l.Name]
	if !known {
		return true, errUnknownNode
	}

	state := l.Sample(ctx)
	if state.OnAC != nil && !*state.OnAC {
		return true, nil
	}

	idle := time.Duration(0)
	if state.Idle != nil {
		idle = *state.Idle
	}

	lease := l.Link.LeaseTask(ctx, &link.TaskRequest{UserActive: idle < node.IdleThreshold, CurrentIdleS: idle.Seconds()})
	if lease == nil {
		return true, nil
	}

	l.runLeased(ctx, cfg, lease)

	return false, nil
}

// runLeased runs one leased task under its profile and completes it, fenced
// by its attempt. A running task is never preempted: it costs quota, and its
// lease is heartbeated to the end; a stopped loop hands it back.
func (l *Loop) runLeased(ctx context.Context, cfg *config.Config, lease *contract.Lease) {
	run := &attempt{
		loop:     l,
		lease:    lease,
		executor: contract.Executor{Node: l.Name, Provider: provider},
		started:  l.Now(),
	}

	var report *link.Completion

	profile, known := cfg.Profiles[lease.Model]

	switch {
	case !known:
		report = run.refuse(unknownProfile)
	case !policy.PrivacyAllows(cfg, lease.Privacy, provider, cfg.Nodes[l.Name].Trust):
		report = run.refuse(privacyRefused)
	default:
		run.profile = &profile
		run.executor.Provider, run.executor.Model = profile.Runner, profile.Model
		report = run.run(ctx)
	}

	if report == nil {
		l.logf("worker: task fenced job=%s\n", lease.JobID)

		return
	}

	if report.Outcome != string(succeeded) {
		// Outcome and code are fixed allowlisted words; the job id is opaque.
		l.logf("worker: task %s: %s job=%s\n", report.Outcome, *report.ErrorCode, lease.JobID)
	}

	l.Link.Complete(ctx, lease.AttemptID, lease.Generation, report)
}

// logf writes one log line; a failed write is not worth stopping for.
func (l *Loop) logf(format string, args ...any) {
	_, _ = fmt.Fprintf(l.Log, format, args...)
}
