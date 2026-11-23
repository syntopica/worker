package client_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/syntopica/worker/internal/client"
	"github.com/syntopica/worker/internal/contract"
)

// exchange is one request the fake coordinator saw.
type exchange struct {
	Method, Path, Query, Token, Body string
}

// record is what the fake coordinator was asked, safe to read while it serves.
type record struct {
	mu   sync.Mutex
	seen []exchange
}

func (r *record) add(e *exchange) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.seen = append(r.seen, *e)
}

func (r *record) all() []exchange {
	r.mu.Lock()
	defer r.mu.Unlock()

	return slices.Clone(r.seen)
}

// coordinator serves answer for every request and records what it was asked.
func coordinator(t *testing.T, status int, answer string) (*client.Client, *record) {
	t.Helper()

	seen := &record{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		seen.add(&exchange{
			Method: r.Method,
			Path:   r.URL.EscapedPath(),
			Query:  r.URL.RawQuery,
			Token:  r.Header.Get("Authorization"),
			Body:   string(body),
		})

		w.WriteHeader(status)
		_, _ = io.WriteString(w, answer)
	}))
	t.Cleanup(server.Close)

	return client.New(server.URL+"/", "token-a", server.Client()), seen
}

func TestClientSubmit(t *testing.T) {
	t.Parallel()

	producer, seen := coordinator(t, http.StatusCreated, `{"id":"j-1","created":true}`)

	job := contract.Submission{Contract: contract.Version, Kind: "task", Queue: "producer-a.refine", Input: json.RawMessage(`{}`)}

	submitted, err := producer.Submit(t.Context(), &job)
	if err != nil {
		t.Fatalf("Submit: %v", err)
	}

	if submitted.ID != "j-1" || !submitted.Created {
		t.Errorf("Submit = %+v, want j-1 created", submitted)
	}

	got := seen.all()[0]
	if got.Method != http.MethodPost || got.Path != "/v1/jobs" || got.Token != "Bearer token-a" {
		t.Errorf("request = %+v", got)
	}

	var sent contract.Submission
	if err = json.Unmarshal([]byte(got.Body), &sent); err != nil || sent.Queue != job.Queue {
		t.Errorf("body = %s (%v), want the submission", got.Body, err)
	}
}

func TestClientRefusal(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct{ body, code string }{
		"coded":      {body: `{"error":"idempotency_conflict"}`, code: "idempotency_conflict"},
		"not json":   {body: `<html>`, code: "unknown"},
		"not object": {body: `[1]`, code: "unknown"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			producer, _ := coordinator(t, http.StatusConflict, test.body)

			_, err := producer.Get(t.Context(), "j-1")

			var refusal *client.Error
			if !errors.As(err, &refusal) {
				t.Fatalf("Get error = %v, want a *client.Error", err)
			}

			if refusal.Status != http.StatusConflict || refusal.Code != test.code {
				t.Errorf("refusal = %v, want 409 %s", refusal, test.code)
			}
		})
	}
}

func TestClientEscapesTheJobID(t *testing.T) {
	t.Parallel()

	producer, seen := coordinator(t, http.StatusOK, `{"state":"cancelled"}`)

	state, err := producer.Cancel(t.Context(), "a/b c")
	if err != nil || state != "cancelled" {
		t.Fatalf("Cancel = %q, %v", state, err)
	}

	if err = producer.Ack(t.Context(), "a/b c", contract.Ack{ResultID: "r-1", Rating: "good"}); err != nil {
		t.Fatalf("Ack: %v", err)
	}

	if got := seen.all()[0].Path; got != "/v1/jobs/a%2Fb%20c/cancel" {
		t.Errorf("cancel path = %s", got)
	}

	if got := seen.all()[1]; got.Path != "/v1/jobs/a%2Fb%20c/ack" || got.Body != `{"result_id":"r-1","decline":false,"rating":"good"}` {
		t.Errorf("ack = %+v", got)
	}
}

func TestClientResults(t *testing.T) {
	t.Parallel()

	producer, seen := coordinator(t, http.StatusOK, `{"results":[{"seq":12,"result_id":"r-1","job_id":"j-1"}]}`)

	results, err := producer.Results(t.Context(), client.ResultsQuery{Queue: "producer-a.bulk", After: 11, Wait: 1500 * time.Millisecond})
	if err != nil {
		t.Fatalf("Results: %v", err)
	}

	if len(results) != 1 || results[0].Seq != 12 {
		t.Errorf("Results = %+v", results)
	}

	if got := seen.all()[0].Query; got != "after=11&limit=50&queue=producer-a.bulk&wait=1.5" {
		t.Errorf("query = %s", got)
	}
}

func TestClientWaitFor(t *testing.T) {
	t.Parallel()

	ready, _ := coordinator(t, http.StatusOK, `{"id":"j-1","state":"succeeded","result":{"seq":1,"result_id":"r-1","job_id":"j-1"}}`)

	result, err := ready.WaitFor(t.Context(), "j-1", time.Millisecond)
	if err != nil || result.ResultID != "r-1" {
		t.Fatalf("WaitFor = %+v, %v", result, err)
	}

	pending, seen := coordinator(t, http.StatusOK, `{"id":"j-1","state":"queued","result":null}`)

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Millisecond)
	defer cancel()

	if _, err = pending.WaitFor(ctx, "j-1", time.Millisecond); !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("WaitFor on a pending job = %v, want the deadline", err)
	}

	if len(seen.all()) < 2 {
		t.Errorf("polled %d times, want more than one", len(seen.all()))
	}
}
