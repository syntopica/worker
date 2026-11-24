package remote

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"
)

// fakeOpenRouter answers /chat/completions with answer (or status) and /key
// with remaining, and records the bodies it was sent.
type fakeOpenRouter struct {
	mu        sync.Mutex
	status    int
	answer    string
	remaining string
	bodies    []map[string]any
}

func newFakeOpenRouter(t *testing.T) (*fakeOpenRouter, *OpenRouter) {
	t.Helper()

	fake := &fakeOpenRouter{
		status:    http.StatusOK,
		remaining: `{"data": {"free_model_daily_requests": {"remaining": 5}}}`,
		answer: `{"choices": [{"message": {"content": "{\"label\": \"x\"}"}}],` +
			` "usage": {"prompt_tokens": 7, "completion_tokens": 3, "cost": 0}}`,
	}
	server := httptest.NewServer(fake)
	t.Cleanup(server.Close)

	return fake, &OpenRouter{API: server.URL, Key: "k", HTTP: server.Client()}
}

func (f *fakeOpenRouter) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if request.Header.Get("Authorization") != "Bearer k" {
		w.WriteHeader(http.StatusUnauthorized)

		return
	}

	if request.Method == http.MethodGet {
		_, _ = w.Write([]byte(f.remaining))

		return
	}

	raw, _ := io.ReadAll(request.Body)

	var body map[string]any

	_ = json.Unmarshal(raw, &body)
	f.bodies = append(f.bodies, body)

	w.WriteHeader(f.status)
	_, _ = w.Write([]byte(f.answer))
}

func (f *fakeOpenRouter) set(status int, answer string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.status, f.answer = status, answer
}

func (f *fakeOpenRouter) sent() []map[string]any {
	f.mu.Lock()
	defer f.mu.Unlock()

	return append([]map[string]any(nil), f.bodies...)
}

func TestPostMapsAnswersAndRefusalsToAllowlistedCodes(t *testing.T) {
	t.Parallel()

	fake, remote := newFakeOpenRouter(t)

	answer, code := remote.Post(t.Context(), map[string]any{"model": "m"})
	if code != "" || answer.Content != `{"label": "x"}` || answer.TokensIn != 7 || answer.TokensOut != 3 {
		t.Fatalf("answer %+v code %q", answer, code)
	}

	refusals := []struct {
		status int
		answer string
		want   string
	}{
		{http.StatusTooManyRequests, `{"error": {}}`, "rate_limited"},
		{http.StatusBadGateway, `{"error": {}}`, "http_502"},
		{http.StatusOK, `{"error": {"message": "provider text"}}`, "bad_response"},
		{http.StatusOK, `{"choices": []}`, "bad_response"},
		{http.StatusOK, `not json`, "bad_response"},
		{http.StatusOK, `{"choices": [{"message": {"content": 3}}]}`, "bad_response"},
	}

	for _, refusal := range refusals {
		fake.set(refusal.status, refusal.answer)

		if answer, code := remote.Post(t.Context(), map[string]any{"model": "m"}); answer != nil || code != refusal.want {
			t.Errorf("%d %s: answer %+v code %q, want %q", refusal.status, refusal.answer, answer, code, refusal.want)
		}
	}

	if _, code := remote.Post(t.Context(), map[string]any{"bad": make(chan int)}); code != "bad_response" {
		t.Errorf("unencodable body: %q", code)
	}
}

func TestAnUnreachableOpenRouterIsATransportErrorAndUnknownHeadroom(t *testing.T) {
	t.Parallel()

	down := &OpenRouter{API: "http://127.0.0.1:1", Key: "k", HTTP: http.DefaultClient}

	if _, code := down.Post(t.Context(), map[string]any{}); code != "transport_error" {
		t.Errorf("post code %q", code)
	}

	if _, known := down.Headroom(t.Context()); known {
		t.Error("headroom known")
	}

	bad := &OpenRouter{API: "://", Key: "k", HTTP: http.DefaultClient}
	if _, code := bad.Post(t.Context(), map[string]any{}); code != "transport_error" {
		t.Errorf("bad url code %q", code)
	}
}

func TestHeadroomReadsTheFreeDailyCounter(t *testing.T) {
	t.Parallel()

	fake, remote := newFakeOpenRouter(t)

	if left, known := remote.Headroom(t.Context()); left != 5 || !known {
		t.Errorf("headroom %d %v", left, known)
	}

	fake.remaining = `{"data": {}}`

	if _, known := remote.Headroom(t.Context()); known {
		t.Error("headroom known without a counter")
	}

	other := &OpenRouter{API: remote.API, Key: "other", HTTP: remote.HTTP}
	if _, known := other.Headroom(t.Context()); known {
		t.Error("headroom known on a refusal")
	}
}

func TestOutputParsesFencedJSONAndReportsCost(t *testing.T) {
	t.Parallel()

	answer, code := decodeAnswer([]byte(`{"model": "vendor/other:free",` +
		` "choices": [{"message": {"content": "` + "```json\\n{\\\"a\\\": 1}\\n```" + `"}}],` +
		` "usage": {"prompt_tokens": 5, "completion_tokens": 2, "cost": 0.25}}`))
	if code != "" || answer.Model != "vendor/other:free" {
		t.Fatalf("answer %+v code %q", answer, code)
	}

	if got := string(as[json.RawMessage](t, answer.Output()["json"])); got != `{"a": 1}` {
		t.Errorf("json %q", got)
	}

	want := map[string]any{"tokens_in": 5, "tokens_out": 2, "cost_usd": 0.25}
	if !reflect.DeepEqual(answer.Usage(), want) {
		t.Errorf("usage %v", answer.Usage())
	}
}
