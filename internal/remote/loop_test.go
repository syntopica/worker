package remote

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
	"github.com/syntopica/worker/internal/link"
)

const (
	testNode  = "node-a"
	testModel = "vendor/model:free"
	testQueue = "pa.bulk"
)

// fakeLink hands out one lease and records what the loop sends.
type fakeLink struct {
	mu        sync.Mutex
	lease     *contract.Lease
	fenced    bool
	completed []*link.Completion
}

func (f *fakeLink) LeaseRemote(context.Context) *contract.Lease {
	f.mu.Lock()
	defer f.mu.Unlock()

	lease := f.lease
	f.lease = nil

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

// fakeProvider answers every call the same way, after wait is closed when set.
type fakeProvider struct {
	mu     sync.Mutex
	left   int
	known  bool
	answer *Answer
	code   string
	wait   chan struct{}
	bodies []map[string]any
}

func (f *fakeProvider) Headroom(context.Context) (left int, known bool) {
	return f.left, f.known
}

func (f *fakeProvider) Post(ctx context.Context, body map[string]any) (answer *Answer, code string) {
	f.mu.Lock()
	f.bodies = append(f.bodies, body)
	f.mu.Unlock()

	if f.wait != nil {
		select {
		case <-f.wait:
		case <-ctx.Done():
			return nil, string(transportError)
		}
	}

	return f.answer, f.code
}

func testLease(privacy string) *contract.Lease {
	return &contract.Lease{
		JobID: "j", AttemptID: "a", Generation: 1, Model: testModel, Queue: testQueue, Privacy: privacy,
		Input: json.RawMessage(`{"messages": [{"role": "user", "content": "hi"}]}`),
	}
}

func testConfig() *config.Config {
	return &config.Config{
		Nodes:   map[string]config.NodePolicy{testNode: {Name: testNode, Trust: "owner"}},
		Queues:  map[string]config.QueuePolicy{testQueue: {Name: testQueue}},
		Privacy: map[string][]string{"public": {provider}, "internal": {"ollama"}},
		Trust:   map[string][]string{"public": {"owner"}, "internal": {"owner"}},
	}
}

func testLoop(coordinator *fakeLink, remote Provider, cfg *config.Config) (*Loop, *bytes.Buffer) {
	log := &bytes.Buffer{}

	return &Loop{
		Name:   testNode,
		Config: func() *config.Config { return cfg },
		Link:   coordinator,
		Remote: remote,
		Now:    time.Now,
		Sleep:  func(context.Context, time.Duration) {},
		Beat:   time.Millisecond,
		Log:    log,
	}, log
}

func TestNoHeadroomOrUnknownHeadroomRestsWithoutLeasing(t *testing.T) {
	t.Parallel()

	coordinator := &fakeLink{lease: testLease("public")}

	for _, remote := range []*fakeProvider{{left: 0, known: true}, {left: 9, known: false}} {
		loop, _ := testLoop(coordinator, remote, testConfig())
		if rest, err := loop.step(t.Context()); rest != noHeadroomRest || err != nil {
			t.Errorf("rest %v err %v", rest, err)
		}
	}

	if coordinator.lease == nil || coordinator.ended() != "" {
		t.Error("the lease was taken")
	}
}

func TestNothingToEscalateRestsBriefly(t *testing.T) {
	t.Parallel()

	loop, _ := testLoop(&fakeLink{}, &fakeProvider{left: 3, known: true}, testConfig())
	if rest, err := loop.step(t.Context()); rest != idleRest || err != nil {
		t.Errorf("rest %v err %v", rest, err)
	}
}

func TestARunReportsTheRemoteExecutorAndTheModelActuallyUsed(t *testing.T) {
	t.Parallel()

	coordinator := &fakeLink{lease: testLease("public")}
	remote := &fakeProvider{left: 3, known: true, answer: &Answer{Model: "vendor/other:free", Content: "{}", TokensIn: 7}}
	loop, log := testLoop(coordinator, remote, testConfig())

	if rest, err := loop.step(t.Context()); rest != minGap || err != nil {
		t.Fatalf("rest %v err %v", rest, err)
	}

	done := coordinator.completed[0]
	if done.Outcome != string(succeeded) || done.ErrorCode != nil || done.Usage["tokens_in"] != 7 {
		t.Errorf("completion %+v", done)
	}

	if done.Executor != (contract.Executor{Node: testNode, Provider: provider, Model: "vendor/other:free"}) {
		t.Errorf("executor %+v", done.Executor)
	}

	if _, zdr := remote.bodies[0]["provider"]; zdr || log.Len() != 0 {
		t.Errorf("public job sent with zdr, or logged: %q", log)
	}
}

func TestARateLimitedCallRestsTheLoopFiveMinutes(t *testing.T) {
	t.Parallel()

	coordinator := &fakeLink{lease: testLease("public")}
	loop, log := testLoop(coordinator, &fakeProvider{left: 3, known: true, code: string(rateLimited)}, testConfig())

	if rest, err := loop.step(t.Context()); rest != rateLimitedRest || err != nil {
		t.Errorf("rest %v err %v", rest, err)
	}

	if coordinator.ended() != "failed/rate_limited" || log.String() != "worker: remote failed: rate_limited job=j\n" {
		t.Errorf("ended %q log %q", coordinator.ended(), log)
	}
}

func TestALeaseThisNodeMayNotSendRemoteIsRefusedUnrun(t *testing.T) {
	t.Parallel()

	coordinator := &fakeLink{lease: testLease("internal")}
	remote := &fakeProvider{left: 3, known: true}
	loop, _ := testLoop(coordinator, remote, testConfig())

	if rest, err := loop.step(t.Context()); rest != minGap || err != nil {
		t.Errorf("rest %v err %v", rest, err)
	}

	if coordinator.ended() != "failed/privacy_refused" || len(remote.bodies) != 0 {
		t.Errorf("ended %q, %d calls", coordinator.ended(), len(remote.bodies))
	}
}

func TestAnUnknownNodeFailsTheIterationAndTheLoopRests(t *testing.T) {
	t.Parallel()

	loop, log := testLoop(&fakeLink{}, &fakeProvider{}, testConfig())
	loop.Name = "node-z"

	ctx, cancel := context.WithCancel(t.Context())

	var rested time.Duration

	loop.Sleep = func(_ context.Context, wait time.Duration) {
		rested = wait

		cancel()
	}

	if err := loop.Run(ctx); err == nil || rested != errorRest {
		t.Errorf("err %v rested %v", err, rested)
	}

	if log.String() != "worker: remote iteration failed: unknown_node\n" {
		t.Errorf("log %q", log)
	}
}

// as is a decoded JSON value of the type the test expects.
func as[T any](t *testing.T, value any) T {
	t.Helper()

	typed, ok := value.(T)
	if !ok {
		t.Fatalf("%v is a %T", value, value)
	}

	return typed
}
