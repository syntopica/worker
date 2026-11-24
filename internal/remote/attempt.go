package remote

import (
	"context"
	"time"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/contract"
	"github.com/syntopica/worker/internal/link"
)

// publicClass is the privacy class that needs no zero data retention.
const publicClass = "public"

// attempt is one leased inference job on its way through the remote endpoint.
type attempt struct {
	loop     *Loop
	lease    *contract.Lease
	executor contract.Executor
	started  time.Time
}

// posted is what a chat completion call came back with.
type posted struct {
	answer *Answer
	code   string
}

// run sends the job and heartbeats its lease while the call runs. The
// completion is reported unless the attempt was fenced out; why is the
// reported error code.
//
// Non-public jobs go to zero-data-retention endpoints unless the queue's
// route turns that off. A pinned job asks its own model only, never the
// route's fallbacks. A stopped loop hands the job back as a shutdown.
func (a *attempt) run(ctx context.Context, cfg *config.Config) (how outcome, why reason) {
	zdr, fallbacks := true, []string(nil)

	if route := cfg.Queues[a.lease.Queue].OpenRouter; route != nil {
		zdr = route.ZDR

		if !a.lease.Pinned {
			fallbacks = route.Fallbacks
		}
	}

	body, err := RequestBody(a.lease.Model, a.lease.Input, zdr && a.lease.Privacy != publicClass, fallbacks)
	if err != nil {
		a.complete(ctx, failed, nodeError, nil)

		return failed, nodeError
	}

	// A fenced or stopped attempt cancels its call.
	callCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	done := make(chan posted, 1)

	go func() {
		answer, code := a.loop.Remote.Post(callCtx, body)
		done <- posted{answer, code}
	}()

	beat := time.NewTicker(a.loop.Beat)
	defer beat.Stop()

	for {
		select {
		case got := <-done:
			return a.settle(ctx, got)
		case <-ctx.Done():
			return a.settle(ctx, posted{})
		case <-beat.C:
			if !a.loop.Link.Heartbeat(ctx, a.lease.AttemptID, a.lease.Generation, false) {
				return fenced, ""
			}
		}
	}
}

// settle reports the attempt's end. A call that did not answer on a stopped
// loop is handed back, which charges nothing; if the coordinator cannot be
// reached the lease simply expires.
func (a *attempt) settle(ctx context.Context, got posted) (how outcome, why reason) {
	switch {
	case got.answer != nil:
		if got.answer.Model != "" {
			a.executor.Model = got.answer.Model
		}

		how = succeeded
	case ctx.Err() != nil:
		how, why = preempted, nodeShutdown
	default:
		how, why = failed, reason(got.code)
	}

	a.complete(ctx, how, why, got.answer)

	return how, why
}

// complete reports how the attempt ended.
func (a *attempt) complete(ctx context.Context, how outcome, why reason, answer *Answer) {
	report := &link.Completion{
		Outcome:  string(how),
		Usage:    map[string]any{},
		Executor: a.executor,
		WallS:    a.loop.Now().Sub(a.started).Seconds(),
	}

	if why != "" {
		report.ErrorCode = new(string(why))
	}

	if answer != nil {
		report.Output, report.Usage = answer.Output(), answer.Usage()
	}

	a.loop.Link.Complete(ctx, a.lease.AttemptID, a.lease.Generation, report)
}
