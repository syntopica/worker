package node

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/contract"
	"github.com/syntopica/worker/internal/host"
	"github.com/syntopica/worker/internal/link"
	"github.com/syntopica/worker/internal/ollama"
	"github.com/syntopica/worker/internal/ollamatest"
)

const (
	testModel = "model-a"
	testNode  = "node-a"
)

// fakeLink is a scripted coordinator that records what the node sends.
type fakeLink struct {
	mu        sync.Mutex
	leases    []*contract.Lease
	fenced    bool
	reasons   []string
	completed []*link.Completion
	beats     []bool
	requests  []link.LeaseRequest
}

func (f *fakeLink) Lease(_ context.Context, request link.LeaseRequest) *contract.Lease {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.requests = append(f.requests, request)
	if len(f.leases) == 0 {
		return nil
	}

	lease := f.leases[0]
	f.leases = f.leases[1:]

	return lease
}

func (f *fakeLink) Heartbeat(_ context.Context, _ string, _ int64, draining bool) bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.beats = append(f.beats, draining)

	return !f.fenced
}

func (f *fakeLink) Complete(_ context.Context, _ string, _ int64, report *link.Completion) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.completed = append(f.completed, report)
}

func (f *fakeLink) Report(_ context.Context, report *link.HostReport) {
	f.mu.Lock()
	defer f.mu.Unlock()

	why := ""
	if report.Reason != nil {
		why = *report.Reason
	}

	f.reasons = append(f.reasons, why)
}

// ended lists each completion as "outcome/error_code".
func (f *fakeLink) ended() []string {
	f.mu.Lock()
	defer f.mu.Unlock()

	out := make([]string, 0, len(f.completed))

	for _, report := range f.completed {
		code := ""
		if report.ErrorCode != nil {
			code = *report.ErrorCode
		}

		out = append(out, report.Outcome+"/"+code)
	}

	return out
}

func hostState(idle time.Duration, onAC bool, pressure host.Pressure) host.State {
	return host.State{Idle: &idle, OnAC: &onAC, Pressure: pressure}
}

func idleState() host.State { return hostState(15*time.Minute, true, host.PressureNormal) }

func typingState() host.State { return hostState(time.Second, true, host.PressureNormal) }

func testLease(runWhen string) *contract.Lease {
	return &contract.Lease{
		JobID: "j", AttemptID: "a", Generation: 1, Model: testModel, Privacy: "mail",
		Input: json.RawMessage(`{"messages": []}`), RunWhen: runWhen,
	}
}

// harness is a node wired to a scripted link, host and clock. The clock
// advances 200 s per reading; a short sleep really waits 20 ms and the n-th
// rest stops the loop.
type harness struct {
	node   *Node
	link   *fakeLink
	policy config.NodePolicy
	log    *bytes.Buffer
	stop   context.CancelFunc
	rests  int
	onRest func(n int)
}

func newHarness(t *testing.T, url string, stopAt int, sample func() host.State) (*harness, context.Context) {
	t.Helper()

	ctx, stop := context.WithCancel(t.Context())
	t.Cleanup(stop)

	h := &harness{link: &fakeLink{}, log: &bytes.Buffer{}, stop: stop}
	h.policy = config.NodePolicy{
		Name: testNode, Trust: "owner", IdleThreshold: 5 * time.Minute, MemoryBudgetGB: 44,
		OllamaURL: url, OllamaLaunchdLabel: "label-a",
	}

	var ticks int64

	h.node = &Node{
		Name: testNode,
		Config: func() *config.Config {
			return &config.Config{
				Models:  map[string]config.ModelPin{testModel: {Name: testModel, NumCtx: 40960, KeepAlive: "5m", ColdGB: 30, WarmGB: 2}},
				Nodes:   map[string]config.NodePolicy{testNode: h.policy},
				Privacy: map[string][]string{"mail": {provider}},
				Trust:   map[string][]string{"mail": {"owner"}},
			}
		},
		Link: h.link,
		Backend: func(url string) ollama.Backend {
			return &ollama.Server{URL: url, HTTP: http.DefaultClient, Run: func(context.Context, string, ...string) (string, bool) {
				return "", false
			}}
		},
		HTTP:   http.DefaultClient,
		Sample: func(context.Context) host.State { return sample() },
		Now: func() time.Time {
			ticks++

			return time.Unix((ticks-1)*200, 0)
		},
		Sleep: func(_ context.Context, wait time.Duration) {
			if wait < restFor {
				time.Sleep(20 * time.Millisecond)

				return
			}

			h.rests++
			if h.onRest != nil {
				h.onRest(h.rests)
			}

			if h.rests >= stopAt {
				stop()
			}
		},
		Log: h.log,
	}

	return h, ctx
}

func equal(t *testing.T, got, want []string) {
	t.Helper()

	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestRunLeasesRunsAndReportsAJob(t *testing.T) {
	t.Parallel()

	fake := ollamatest.New(t)
	h, ctx := newHarness(t, fake.URL, 2, idleState)
	h.link.leases = []*contract.Lease{testLease(runWhenIdle)}

	_ = h.node.Run(ctx)

	equal(t, h.link.reasons, []string{string(pressureRecovering), "", ""})
	equal(t, h.link.ended(), []string{"succeeded/"})

	done := h.link.completed[0]
	if done.Output["text"] != `{"label": "x"}` || done.Usage["tokens_in"] != 7 || done.WallS <= 0 {
		t.Fatalf("completion: %+v", done)
	}

	request := h.link.requests[0]
	if request.ResidentModel == nil || *request.ResidentModel != testModel || request.FreeGB != 14 || request.UserActive {
		t.Fatalf("lease request: %+v", request)
	}
}

func TestRunNeverLeasesOnBatteryOrWithTheBackendDown(t *testing.T) {
	t.Parallel()

	battery, ctx := newHarness(t, ollamatest.New(t).URL, 2, func() host.State {
		return hostState(15*time.Minute, false, host.PressureNormal)
	})
	_ = battery.node.Run(ctx)

	down, ctx := newHarness(t, "http://127.0.0.1:1", 3, idleState)
	_ = down.node.Run(ctx)

	equal(t, battery.link.reasons, []string{string(onBattery), string(onBattery)})
	equal(t, down.link.reasons[1:], []string{string(backendDown), string(backendDown)})

	if len(battery.link.requests)+len(down.link.requests) != 0 {
		t.Fatal("leased while blocked")
	}
}

func TestRunRefusesALeaseItMayNotRun(t *testing.T) {
	t.Parallel()

	unknown, secret := testLease(runWhenIdle), testLease(runWhenIdle)
	unknown.Model, secret.Privacy = "model-zz", "secret"

	h, ctx := newHarness(t, ollamatest.New(t).URL, 2, idleState)
	h.link.leases = []*contract.Lease{unknown, secret}

	_ = h.node.Run(ctx)

	equal(t, h.link.ended(), []string{"failed/unknown_model", "failed/privacy_refused"})
}

func TestRunReportsABadInputAsNodeErrorWithoutLoggingIt(t *testing.T) {
	t.Parallel()

	bad := testLease(runWhenIdle)
	bad.Input = json.RawMessage(`["generated text"]`)

	h, ctx := newHarness(t, ollamatest.New(t).URL, 3, idleState)
	h.link.leases = []*contract.Lease{bad}

	_ = h.node.Run(ctx)

	equal(t, h.link.ended(), []string{"failed/node_error"})

	logged := h.log.String()
	if !strings.Contains(logged, "worker: attempt failed: bad_input\n") ||
		!strings.Contains(logged, "worker: attempt failed: node_error job=j\n") || strings.Contains(logged, "generated") {
		t.Fatalf("log: %q", logged)
	}
}

func TestRunLogsAnUnknownNodeAndKeepsGoing(t *testing.T) {
	t.Parallel()

	h, ctx := newHarness(t, "http://127.0.0.1:1", 2, idleState)
	h.node.Name = "node-zz"

	if err := h.node.Run(ctx); err == nil || h.rests != 2 {
		t.Fatalf("err %v after %d rests", err, h.rests)
	}

	if !strings.Contains(h.log.String(), "worker: node iteration failed: unknown_node\n") {
		t.Fatalf("log: %q", h.log.String())
	}
}

func TestAFailedDrainBlocksUntilTheModelIsUnlisted(t *testing.T) {
	t.Parallel()

	fake := ollamatest.New(t)
	h, ctx := newHarness(t, fake.URL, 3, idleState)
	h.node.remembered().failedModel = testModel
	h.onRest = func(n int) {
		if n == 2 {
			fake.Set(func(f *ollamatest.Fake) { f.PsRaw = `{"models": [{"name": "model-a"}]}` })
		}
	}

	_ = h.node.Run(ctx)

	// The blocked step retries the unload, which empties the fake.
	equal(t, h.link.reasons, []string{string(drainBlocked), "", ""})

	for _, body := range fake.Bodies() {
		if options, _ := body["options"].(map[string]any); options["num_predict"] != nil {
			t.Fatal("a failed drain was probed")
		}
	}
}

func TestALocalIdleOnlyNodeNeitherLeasesNorKeepsItsModelWhileActive(t *testing.T) {
	t.Parallel()

	fake := ollamatest.New(t)
	h, ctx := newHarness(t, fake.URL, 1, func() host.State { return hostState(30*time.Second, true, host.PressureNormal) })
	h.policy.LocalIdleOnly = true
	h.node.remembered().owned[testModel] = true
	h.node.remembered().normalSince = new(time.Unix(-1000, 0))

	_ = h.node.Run(ctx)

	if len(h.link.requests) != 0 || len(fake.Bodies()) != 1 || fake.Bodies()[0]["keep_alive"] != 0.0 {
		t.Fatalf("requests %d, bodies %v", len(h.link.requests), fake.Bodies())
	}
}

func TestSustainedPressureUnloadsOnlyAModelTheNodeLoaded(t *testing.T) {
	t.Parallel()

	fake := ollamatest.New(t)
	critical := func() host.State { return hostState(15*time.Minute, true, host.PressureCritical) }

	startup, ctx := newHarness(t, fake.URL, 5, critical)
	_ = startup.node.Run(ctx)

	if len(fake.Bodies()) != 0 || startup.link.reasons[4] != string(memoryPressure) {
		t.Fatalf("bodies %v, reasons %v", fake.Bodies(), startup.link.reasons)
	}

	owned, ctx := newHarness(t, fake.URL, 5, critical)
	owned.node.remembered().owned[testModel] = true
	_ = owned.node.Run(ctx)

	if len(fake.Bodies()) != 1 {
		t.Fatalf("unloads after three samples, once per episode: %v", fake.Bodies())
	}
}
