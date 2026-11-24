package remote

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/contract"
)

// ran runs one lease straight through an attempt.
func ran(ctx context.Context, coordinator *fakeLink, remote Provider, cfg *config.Config, privacy string, pinned bool) outcome {
	loop, _ := testLoop(coordinator, remote, cfg)
	lease := testLease(privacy)
	lease.Pinned = pinned

	run := &attempt{
		loop:     loop,
		lease:    lease,
		executor: contract.Executor{Node: testNode, Provider: provider, Model: lease.Model},
		started:  time.Now(),
	}

	how, _ := run.run(ctx, cfg)

	return how
}

func routed() *config.Config {
	cfg := testConfig()
	cfg.Queues[testQueue] = config.QueuePolicy{
		Name:       testQueue,
		OpenRouter: &config.OpenRouterRoute{ZDR: true, Fallbacks: []string{"vendor/other:free"}},
	}

	return cfg
}

func TestANonPublicJobGoesToZeroDataRetentionEndpoints(t *testing.T) {
	t.Parallel()

	fake, remote := newFakeOpenRouter(t)
	coordinator := &fakeLink{}

	if how := ran(t.Context(), coordinator, remote, testConfig(), "internal", false); how != succeeded {
		t.Fatalf("outcome %q", how)
	}

	done := coordinator.completed[0]
	if string(as[json.RawMessage](t, done.Output["json"])) != `{"label": "x"}` || done.Executor.Provider != provider {
		t.Errorf("completion %+v", done)
	}

	if zdr := as[map[string]any](t, fake.sent()[0]["provider"])["zdr"]; zdr != true {
		t.Errorf("zdr %v", zdr)
	}
}

func TestTheQueueRouteDecidesZDRAndFallbacks(t *testing.T) {
	t.Parallel()

	sent := func(cfg *config.Config, privacy string, pinned bool) map[string]any {
		remote := &fakeProvider{code: string(transportError)}
		ran(t.Context(), &fakeLink{}, remote, cfg, privacy, pinned)

		return remote.bodies[0]
	}

	if body := sent(routed(), "internal", false); body["provider"] == nil || len(as[[]string](t, body["models"])) != 2 {
		t.Errorf("routed body %v", body)
	}

	if _, set := sent(routed(), "internal", true)["models"]; set {
		t.Error("a pinned job was sent the route's fallbacks")
	}

	off := routed()
	off.Queues[testQueue].OpenRouter.ZDR = false

	for name, body := range map[string]map[string]any{
		"zdr off": sent(off, "internal", false),
		"public":  sent(routed(), "public", false),
	} {
		if _, set := body["provider"]; set {
			t.Errorf("%s: sent with zdr", name)
		}
	}
}

func TestAFencedRemoteAttemptIsAbandoned(t *testing.T) {
	t.Parallel()

	coordinator := &fakeLink{fenced: true}
	remote := &fakeProvider{wait: make(chan struct{})}

	if how := ran(t.Context(), coordinator, remote, testConfig(), "public", false); how != fenced {
		t.Errorf("outcome %q", how)
	}

	if coordinator.ended() != "" {
		t.Errorf("completed %q", coordinator.ended())
	}
}

func TestAStoppingSlotHandsItsRunningCallBack(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	coordinator := &fakeLink{}
	remote := &fakeProvider{wait: make(chan struct{})}

	if how := ran(ctx, coordinator, remote, testConfig(), "public", false); how != preempted {
		t.Errorf("outcome %q", how)
	}

	if coordinator.ended() != "preempted/node_shutdown" {
		t.Errorf("completed %q", coordinator.ended())
	}
}

func TestAnInputTheEndpointCannotBeSentFailsTheAttempt(t *testing.T) {
	t.Parallel()

	coordinator := &fakeLink{}
	loop, _ := testLoop(coordinator, &fakeProvider{}, testConfig())
	lease := testLease("public")
	lease.Input = json.RawMessage(`{}`)

	if how, why := (&attempt{loop: loop, lease: lease}).run(t.Context(), testConfig()); how != failed || why != nodeError {
		t.Errorf("outcome %q %q", how, why)
	}

	if coordinator.ended() != "failed/node_error" {
		t.Errorf("completed %q", coordinator.ended())
	}
}

func TestEverySlotRunsAndStopsWithTheContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(t.Context())
	stepped := make(chan struct{}, 3)

	slots := make([]*Loop, 3)
	for at := range slots {
		slots[at], _ = testLoop(&fakeLink{}, &fakeProvider{}, testConfig())
		slots[at].Sleep = func(ctx context.Context, _ time.Duration) {
			stepped <- struct{}{}

			<-ctx.Done()
		}
	}

	go func() {
		for range slots {
			<-stepped
		}

		cancel()
	}()

	if err := RunSlots(ctx, slots); err == nil {
		t.Error("no error from stopped slots")
	}
}
