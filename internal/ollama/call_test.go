package ollama_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/syntopica/worker/internal/ollama"
	"github.com/syntopica/worker/internal/ollamatest"
)

const idle = 30 * time.Second

// start sends the standard streamed chat to the fake.
func start(ctx context.Context, fake *ollamatest.Fake, idle time.Duration) *ollama.Call {
	return ollama.StartCall(ctx, http.DefaultClient, fake.URL, map[string]any{"model": "model-a", "stream": true}, idle)
}

// waitPhase polls until the call is in a phase or the deadline passes.
func waitPhase(call *ollama.Call, phase string) bool {
	for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline); time.Sleep(5 * time.Millisecond) {
		if call.Phase() == phase {
			return true
		}
	}

	return false
}

func TestCallReturnsTheJoinedAnswerAndItsCounts(t *testing.T) {
	t.Parallel()

	call := start(t.Context(), ollamatest.New(t), idle)

	answer, code := call.Outcome()
	if code != "" || answer == nil {
		t.Fatalf("outcome = %v, %q", answer, code)
	}

	if *answer != (ollama.Answer{Content: `{"label": "x"}`, TokensIn: 7, TokensOut: 3}) {
		t.Errorf("answer = %+v", *answer)
	}

	if !call.Done() || call.Phase() != "generation" {
		t.Errorf("done = %v, phase = %s", call.Done(), call.Phase())
	}
}

func TestCallPhaseTurnsToGenerationWithTheFirstChunk(t *testing.T) {
	t.Parallel()

	fake := ollamatest.New(t)
	fake.Delay, fake.GenerationDelay = 200*time.Millisecond, time.Minute
	call := start(t.Context(), fake, idle)

	if call.Phase() != "prefill" || call.Done() {
		t.Fatalf("phase = %s, done = %v before the first chunk", call.Phase(), call.Done())
	}

	if !waitPhase(call, "generation") {
		t.Fatal("the phase never turned to generation")
	}

	call.Cancel()
	<-call.Finished()

	if answer, code := call.Outcome(); answer != nil || code != "cancelled" {
		t.Errorf("outcome = %v, %q, want cancelled", answer, code)
	}
}

func TestCallCancelledBeforeItStartsSendsNothing(t *testing.T) {
	t.Parallel()

	fake := ollamatest.New(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	call := start(ctx, fake, idle)
	call.Cancel()

	if answer, code := call.Outcome(); answer != nil || code != "cancelled" {
		t.Errorf("outcome = %v, %q, want cancelled", answer, code)
	}

	if len(fake.Bodies()) != 0 {
		t.Errorf("the fake received %d requests", len(fake.Bodies()))
	}
}

func TestCallErrorCodesAreAllowlisted(t *testing.T) {
	t.Parallel()

	cases := []struct {
		status int
		raw    string
		want   string
	}{
		{500, "boom secret text", "http_500"},
		{200, `{"error": "x"}`, "bad_response"},
		{200, "not json", "bad_response"},
		{200, "[1]", "bad_response"},
		{200, "null", "bad_response"},
		{200, `{"message": {"content": "a"}, "done": false}` + "\n\n", "bad_response"},
	}

	for _, c := range cases {
		fake := ollamatest.New(t)
		fake.Status, fake.RawBody = c.status, c.raw

		if answer, code := start(t.Context(), fake, idle).Outcome(); answer != nil || code != c.want {
			t.Errorf("%d %q: outcome = %v, %q, want %q", c.status, c.raw, answer, code, c.want)
		}
	}
}

func TestCallToleratesOddChunks(t *testing.T) {
	t.Parallel()

	fake := ollamatest.New(t)
	fake.RawBody = `{"message": {"content": "a"}}` + "\n" + `{"message": {"content": 5}}` + "\n" +
		`{"message": "b"}` + "\n" + `{"message": {"content": "c"}, "done": true, "eval_count": "many"}` + "\n"

	answer, code := start(t.Context(), fake, idle).Outcome()
	if code != "" || answer == nil || *answer != (ollama.Answer{Content: "ac"}) {
		t.Errorf("outcome = %+v, %q", answer, code)
	}
}

func TestCallFailsAsTransportError(t *testing.T) {
	t.Parallel()

	silent := ollamatest.New(t)
	silent.Delay = time.Minute

	if answer, code := start(t.Context(), silent, 50*time.Millisecond).Outcome(); answer != nil || code != "transport_error" {
		t.Errorf("a silent server: outcome = %v, %q", answer, code)
	}

	stalled := ollamatest.New(t)
	stalled.GenerationDelay = time.Minute

	if answer, code := start(t.Context(), stalled, 50*time.Millisecond).Outcome(); answer != nil || code != "transport_error" {
		t.Errorf("a stalled stream: outcome = %v, %q", answer, code)
	}

	unreachable := &ollamatest.Fake{URL: "http://127.0.0.1:1"}

	if answer, code := start(t.Context(), unreachable, idle).Outcome(); answer != nil || code != "transport_error" {
		t.Errorf("an unreachable server: outcome = %v, %q", answer, code)
	}

	body := map[string]any{"unencodable": make(chan int)}

	if _, code := ollama.StartCall(t.Context(), http.DefaultClient, silent.URL, body, idle).Outcome(); code != "transport_error" {
		t.Errorf("an unencodable body: code = %q", code)
	}

	if _, code := ollama.StartCall(t.Context(), http.DefaultClient, "://bad", nil, idle).Outcome(); code != "transport_error" {
		t.Errorf("a bad URL: code = %q", code)
	}
}
