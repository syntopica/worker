package node

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/contract"
	"github.com/syntopica/worker/internal/host"
	"github.com/syntopica/worker/internal/link"
	"github.com/syntopica/worker/internal/ollama"
)

const (
	// sampleEvery is how often a running attempt heartbeats and samples.
	sampleEvery = 2 * time.Second
	// callSilence is the longest silence tolerated from the model server.
	callSilence = 30 * time.Minute
)

// errBadInput is a lease whose input the model server cannot be sent.
var errBadInput = errors.New("bad_input")

// attempt is one leased inference job on its way through the model server.
type attempt struct {
	node     *Node
	policy   *config.NodePolicy
	pin      *config.ModelPin
	lease    *contract.Lease
	backend  ollama.Backend
	executor contract.Executor
	started  time.Time
	pressed  int
}

// run executes the job, releasing it the moment the host says so. The
// completion is reported unless the attempt was fenced out; why is the
// reported error code. Pressure releases only once sustained.
func (a *attempt) run(ctx context.Context) (how outcome, why reason, err error) {
	body, err := ollama.RequestBody(a.pin, a.lease.Input)
	if err != nil {
		return failed, nodeError, errBadInput
	}

	a.started = a.node.Now()

	call := ollama.StartCall(ctx, a.node.HTTP, a.policy.OllamaURL, body, callSilence)
	// Left alone the call would hold the backend and the next lease would
	// queue behind it.
	defer call.Cancel()

	for !call.Done() {
		a.node.Sleep(ctx, sampleEvery)

		if how, why = a.watch(ctx, call); how != "" {
			return how, why, nil
		}
	}

	answer, code := call.Outcome()
	if answer == nil && ctx.Err() != nil {
		a.complete(ctx, preempted, nodeShutdown, nil, a.wall())
		return preempted, nodeShutdown, nil
	}
	if answer == nil {
		a.complete(ctx, failed, reason(code), nil, a.wall())

		return failed, reason(code), nil
	}

	a.complete(ctx, succeeded, "", answer, a.wall())

	return succeeded, "", nil
}

// watch is one look at a running attempt; a non-empty outcome ends it.
func (a *attempt) watch(ctx context.Context, call *ollama.Call) (how outcome, why reason) {
	if ctx.Err() != nil {
		// Stopped: hand the job back, which charges nothing. If the
		// coordinator cannot be reached the lease simply expires.
		return a.stop(ctx, call)
	}

	if call.Done() {
		return "", ""
	}

	if !a.node.Link.Heartbeat(ctx, a.lease.AttemptID, a.lease.Generation, false) {
		call.Cancel()

		return a.drained(ollama.Drain(ctx, a.backend, a.pin, a.policy.OllamaLaunchdLabel, fencedOut), fenced), ""
	}

	if call.Done() {
		return "", ""
	}

	state := a.node.Sample(ctx)
	if ctx.Err() != nil {
		return a.stop(ctx, call)
	}
	if state.Pressure.Pressed() {
		a.pressed++
	} else {
		a.pressed = 0
	}

	runWhen := a.lease.RunWhen
	if a.policy.LocalIdleOnly {
		runWhen = runWhenIdle
	}

	why = sustainedReleaseReason(state, runWhen, a.pressed)
	if why == "" || call.Done() {
		return "", ""
	}

	return a.release(ctx, call, state, why)
}

// release cancels the call, drains the backend and reports the preemption.
func (a *attempt) release(ctx context.Context, call *ollama.Call, state host.State, why reason) (outcome, reason) {
	phase := call.Phase()
	call.Cancel()

	cancelled := a.node.Now()
	quiet := ollama.Drain(ctx, a.backend, a.pin, a.policy.OllamaLaunchdLabel, func(ctx context.Context) bool {
		return a.node.Link.Heartbeat(ctx, a.lease.AttemptID, a.lease.Generation, true)
	})

	if ctx.Err() != nil {
		return a.stop(ctx, call)
	}

	result, load := "failed", ""
	if quiet {
		result = "quiet"
	}

	if state.Load1m != nil {
		load = fmt.Sprintf(" load=%.1f", *state.Load1m)
	}

	// The phase 1c cancel-to-quiet measurement, taken from real preemptions.
	now := a.node.Now()
	a.node.logf("worker: drain %s in %.1fs phase=%s reason=%s at=%s%s\n",
		result, now.Sub(cancelled).Seconds(), phase, why, now.UTC().Format(stampLayout), load)
	a.complete(ctx, preempted, why, nil, a.wall())

	return a.drained(quiet, preempted), why
}

func (a *attempt) stop(ctx context.Context, call *ollama.Call) (outcome, reason) {
	call.Cancel()
	a.complete(ctx, preempted, nodeShutdown, nil, a.wall())
	return preempted, nodeShutdown
}

// complete reports how the attempt ended.
func (a *attempt) complete(ctx context.Context, how outcome, why reason, answer *ollama.Answer, wall float64) {
	report := &link.Completion{
		Outcome:   string(how),
		Usage:     map[string]any{},
		Executor:  a.executor,
		ErrorCode: why.pointer(),
		WallS:     wall,
	}

	if answer != nil {
		report.Output = answer.Output()
		report.Usage = map[string]any{"tokens_in": answer.TokensIn, "tokens_out": answer.TokensOut}
	}

	a.node.Link.Complete(ctx, a.lease.AttemptID, a.lease.Generation, report)
}

// drained is the outcome after a drain: the given one when the backend is
// quiet, a failed drain otherwise.
func (*attempt) drained(quiet bool, how outcome) outcome {
	if quiet {
		return how
	}

	return drainFailed
}

// wall is the time the attempt has taken, in seconds.
func (a *attempt) wall() float64 {
	return a.node.Now().Sub(a.started).Seconds()
}

// fencedOut is the heartbeat of a drain whose lease is already lost: it
// keeps the probes going and nothing more.
func fencedOut(context.Context) bool {
	return true
}
