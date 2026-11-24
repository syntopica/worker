package tasks

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/contract"
	"github.com/syntopica/worker/internal/host"
	"github.com/syntopica/worker/internal/link"
)

const (
	testNode    = "node-a"
	testProfile = "p"
	// codexAnswers writes "{}" to the file codex is told to answer in.
	codexAnswers = `while [ $# -gt 1 ]; do [ "$1" = "-o" ] && printf '{}' > "$2"; shift; done`
)

// fakeLink hands out its leases in order and records what the loop sends.
type fakeLink struct {
	mu        sync.Mutex
	leases    []*contract.Lease
	fenced    bool
	requests  []link.TaskRequest
	completed []*link.Completion
	walls     []map[string]float64
	flushed   int
	onLease   func()
}

func (f *fakeLink) LeaseTask(_ context.Context, request *link.TaskRequest) *contract.Lease {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.requests = append(f.requests, *request)

	if f.onLease != nil {
		f.onLease()
	}

	if len(f.leases) == 0 {
		return nil
	}

	lease := f.leases[0]
	f.leases = f.leases[1:]

	return lease
}

func (f *fakeLink) Heartbeat(context.Context, string, int64, bool) bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	return !f.fenced
}

func (f *fakeLink) Complete(_ context.Context, _ string, _ int64, report *link.Completion) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.completed = append(f.completed, report)
}

func (f *fakeLink) ReportWalls(_ context.Context, walls map[string]float64) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.walls = append(f.walls, walls)
}

func (f *fakeLink) Flush(context.Context) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.flushed++
}

// ended lists each completion as "outcome/error_code".
func (f *fakeLink) ended() string {
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

	return strings.Join(out, " ")
}

func testLease(profile string) *contract.Lease {
	return &contract.Lease{
		JobID: "j", AttemptID: "a1", Generation: 2, Model: profile, Privacy: "internal",
		Input: json.RawMessage(`{"prompt": "x"}`),
	}
}

func testConfig(profile *config.TaskProfile) *config.Config {
	return &config.Config{
		Nodes:    map[string]config.NodePolicy{testNode: {Name: testNode, Trust: "owner", IdleThreshold: 5 * time.Minute}},
		Profiles: map[string]config.TaskProfile{testProfile: *profile},
		Privacy:  map[string][]string{"internal": {provider}, "secret": {"ollama"}},
		Trust:    map[string][]string{"internal": {"owner"}, "secret": {"owner"}},
	}
}

// hostState is a host reading; a negative idle or a nil power is unreadable.
func hostState(idle time.Duration, onAC *bool) host.Sampler {
	state := host.State{OnAC: onAC}
	if idle >= 0 {
		state.Idle = &idle
	}

	return func(context.Context) host.State { return state }
}

func testLoop(coordinator *fakeLink, cfg *config.Config) (*Loop, *bytes.Buffer) {
	log := &bytes.Buffer{}

	return &Loop{
		Name:   testNode,
		Config: func() *config.Config { return cfg },
		Link:   coordinator,
		Sample: hostState(15*time.Minute, new(true)),
		Now:    time.Now,
		Sleep:  func(context.Context, time.Duration) {},
		Beat:   5 * time.Millisecond,
		Log:    log,
	}, log
}

func TestATaskIsLeasedRunAndCompleted(t *testing.T) {
	t.Parallel()

	coordinator := &fakeLink{leases: []*contract.Lease{testLease(testProfile)}}
	loop, _ := testLoop(coordinator, testConfig(fake(t, "codex", codexAnswers)))
	loop.Sample = hostState(5*time.Second, new(true))

	resting, err := loop.step(t.Context(), loop.Config())
	if resting || err != nil {
		t.Fatalf("resting %v, err %v", resting, err)
	}

	// Idle 5 s is below the threshold: the user is active.
	if request := coordinator.requests[0]; !request.UserActive || request.CurrentIdleS != 5 {
		t.Errorf("request %+v", request)
	}

	if got := coordinator.ended(); got != "succeeded/" {
		t.Errorf("ended %q", got)
	}

	if report := coordinator.completed[0]; report.Executor != (contract.Executor{Node: testNode, Provider: "codex", Model: "m-1"}) {
		t.Errorf("executor %+v", report.Executor)
	}
}

func TestOnlyAKnownBatteryTakesNoTask(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		sample     host.Sampler
		leases     int
		userActive bool
	}{
		"battery":          {hostState(15*time.Minute, new(false)), 0, false},
		"unreadable idle":  {hostState(-1, new(true)), 1, true},
		"unreadable power": {hostState(15*time.Minute, nil), 1, false},
		"unreadable host":  {hostState(-1, nil), 1, true},
		"idle on mains":    {hostState(15*time.Minute, new(true)), 1, false},
	}

	for name, test := range cases {
		coordinator := &fakeLink{}
		loop, _ := testLoop(coordinator, testConfig(fake(t, "codex", "")))
		loop.Sample = test.sample

		resting, err := loop.step(t.Context(), loop.Config())
		if !resting || err != nil || len(coordinator.requests) != test.leases {
			t.Errorf("%s: resting %v, err %v, requests %d", name, resting, err, len(coordinator.requests))
		}

		if test.leases == 1 && coordinator.requests[0].UserActive != test.userActive {
			t.Errorf("%s: user active %v", name, coordinator.requests[0].UserActive)
		}
	}
}

func TestALeaseTheNodeMayNotRunIsRefusedWithoutARunner(t *testing.T) {
	t.Parallel()

	secret := testLease(testProfile)
	secret.Privacy = "secret"

	coordinator := &fakeLink{leases: []*contract.Lease{testLease("gone"), secret}}
	loop, log := testLoop(coordinator, testConfig(fake(t, "codex", codexAnswers)))

	for range 2 {
		if resting, err := loop.step(t.Context(), loop.Config()); resting || err != nil {
			t.Fatalf("resting %v, err %v", resting, err)
		}
	}

	if got := coordinator.ended(); got != "failed/unknown_profile failed/privacy_refused" {
		t.Errorf("ended %q", got)
	}

	if report := coordinator.completed[0]; report.WallS != 0 || report.Executor.Provider != provider {
		t.Errorf("report %+v", report)
	}

	if !strings.Contains(log.String(), "worker: task failed: unknown_profile job=j\n") {
		t.Errorf("log %q", log.String())
	}
}

func TestAnUnknownNodeRestsAndIsLogged(t *testing.T) {
	t.Parallel()

	coordinator := &fakeLink{leases: []*contract.Lease{testLease(testProfile)}}
	loop, log := testLoop(coordinator, testConfig(fake(t, "codex", "")))
	loop.Name = "node-z"

	ctx, cancel := context.WithCancel(t.Context())
	loop.Sleep = func(context.Context, time.Duration) { cancel() }

	if err := loop.Run(ctx); err == nil || len(coordinator.requests) != 0 {
		t.Errorf("err %v, requests %d", err, len(coordinator.requests))
	}

	if log.String() != "worker: task iteration failed: unknown_node\n" {
		t.Errorf("log %q", log.String())
	}
}

func TestSlotsRunUntilStoppedAndOnlyAProbingSlotReportsWalls(t *testing.T) {
	t.Parallel()

	cfg := testConfig(fake(t, "codex", ""))
	cfg.RunnerQuota = map[string]config.RunnerQuota{"codex": {Provider: "codex"}}

	ctx, cancel := context.WithCancel(t.Context())
	probing, quiet := &fakeLink{onLease: cancel}, &fakeLink{}

	first, _ := testLoop(probing, cfg)
	first.Now = func() time.Time { return time.Unix(0, 0) }
	first.Usage = func(context.Context, string) map[string]any {
		return map[string]any{"primary": map[string]any{"usedPercent": 100.0, "resetsAt": later}}
	}

	second, _ := testLoop(quiet, cfg)

	if err := RunSlots(ctx, []*Loop{first, second}); err == nil {
		t.Error("stopped slots return the context's error")
	}

	if len(probing.walls) == 0 || probing.walls[0]["codex"] != laterTS || len(quiet.walls) != 0 {
		t.Errorf("walls %v %v", probing.walls, quiet.walls)
	}
}
