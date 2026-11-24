package node

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/syntopica/worker/internal/host"
	"github.com/syntopica/worker/internal/ollamatest"
)

// attemptOn runs one lease through the loop and returns the harness.
func attemptOn(t *testing.T, delay time.Duration, runWhen string, change func(*harness), sample func() host.State) *harness {
	t.Helper()

	fake := ollamatest.New(t)
	fake.Delay = delay

	h, ctx := newHarness(t, fake.URL, 2, idleState)
	h.node.remembered().normalSince = new(time.Unix(-1000, 0))
	h.link.leases = append(h.link.leases, testLease(runWhen))

	leased := false
	h.node.Sample = func(context.Context) host.State {
		if !leased {
			leased = true

			return idleState()
		}

		return sample()
	}

	if change != nil {
		change(h)
	}

	_ = h.node.Run(ctx)

	return h
}

func TestAttemptOutcomes(t *testing.T) {
	t.Parallel()

	warn := func() host.State { return hostState(15*time.Minute, true, host.PressureWarn) }
	flicker := 0
	cases := map[string]struct {
		delay   time.Duration
		runWhen string
		change  func(*harness)
		sample  func() host.State
		want    string
	}{
		"a user return preempts":          {3 * time.Second, runWhenIdle, nil, typingState, "preempted/user_active"},
		"active_ok runs while typing":     {200 * time.Millisecond, "active_ok", nil, typingState, "succeeded/"},
		"local idle only releases anyway": {3 * time.Second, "active_ok", func(h *harness) { h.policy.LocalIdleOnly = true }, typingState, "preempted/user_active"},
		"sustained pressure releases":     {3 * time.Second, runWhenIdle, nil, warn, "preempted/memory_pressure"},
		"intermittent pressure does not": {300 * time.Millisecond, runWhenIdle, nil, func() host.State {
			flicker++
			if flicker%3 == 0 {
				return idleState()
			}

			return warn()
		}, "succeeded/"},
		"a fenced attempt reports nothing": {3 * time.Second, runWhenIdle, func(h *harness) { h.link.fenced = true }, idleState, ""},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			h := attemptOn(t, tc.delay, tc.runWhen, tc.change, tc.sample)

			want := []string{tc.want}
			if tc.want == "" {
				want = nil
			}

			equal(t, h.link.ended(), want)
		})
	}
}

func TestAPreemptionDrainsLogsAndBacksOff(t *testing.T) {
	t.Parallel()

	samples := 0
	h := attemptOn(t, 3*time.Second, runWhenIdle, nil, func() host.State {
		samples++
		state := hostState(15*time.Minute, true, host.PressureWarn)
		state.Load1m = new(17.2)

		return state
	})

	line := regexp.MustCompile(`worker: drain quiet in \d+\.\ds phase=prefill reason=memory_pressure at=\d{4}-\d\d-\d\dT\d\d:\d\d:\d\dZ load=17\.2\n` +
		`worker: attempt preempted: memory_pressure job=j\n`)
	if !line.MatchString(h.log.String()) || samples < pressureSamples {
		t.Fatalf("log %q after %d samples", h.log.String(), samples)
	}

	m := h.node.remembered()
	if m.backoff != 2*firstBackoff {
		t.Fatalf("memory: %+v", m)
	}

	if !h.link.beats[len(h.link.beats)-1] {
		t.Fatal("the drain did not heartbeat as draining")
	}
}

func TestAHTTPErrorFailsTheAttempt(t *testing.T) {
	t.Parallel()

	fake := ollamatest.New(t)
	fake.Status, fake.RawBody = 500, "{}"

	failing, ctx := newHarness(t, fake.URL, 2, idleState)
	failing.node.remembered().normalSince = new(time.Unix(-1000, 0))
	failing.link.leases = append(failing.link.leases, testLease(runWhenIdle))
	_ = failing.node.Run(ctx)

	equal(t, failing.link.ended(), []string{"failed/http_500"})
}

func TestAStoppedNodeHandsItsJobBack(t *testing.T) {
	t.Parallel()

	h := attemptOn(t, 5*time.Second, runWhenIdle, func(h *harness) {
		h.node.Sleep = func(context.Context, time.Duration) { h.stop() }
	}, idleState)

	equal(t, h.link.ended(), []string{"preempted/node_shutdown"})

	if h.log.Len() != 0 {
		t.Fatalf("log: %q", h.log.String())
	}
}

func TestShutdownDuringSamplingDoesNotChargeAPreemption(t *testing.T) {
	t.Parallel()
	h := attemptOn(t, 5*time.Second, runWhenIdle, func(h *harness) {
		sample := h.node.Sample
		first := true
		h.node.Sample = func(ctx context.Context) host.State {
			if first {
				first = false
				return sample(ctx)
			}
			h.stop()
			return host.State{}
		}
	}, idleState)
	equal(t, h.link.ended(), []string{"preempted/node_shutdown"})
}
