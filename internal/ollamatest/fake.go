package ollamatest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// Fake is a model server that records request bodies. A streamed chat
// answers in two NDJSON chunks. Read and change its fields through Set once
// requests may be running.
type Fake struct {
	// URL is the server's base URL.
	URL string
	// Delay slows prefill: the wait before a chat answers.
	Delay time.Duration
	// GenerationDelay is the wait between the two chunks.
	GenerationDelay time.Duration
	// BusyUntil delays probes until then.
	BusyUntil time.Time
	// Loaded is what /api/ps lists; an unload empties it.
	Loaded []string
	// Status is the status of a RawBody answer.
	Status int
	// RawBody, when set, answers every chat verbatim.
	RawBody string
	// PsRaw, when set, answers /api/ps verbatim.
	PsRaw string

	mu     sync.Mutex
	bodies []map[string]any
}

// New starts a fake holding "model-a" and stops it when the test ends.
func New(t *testing.T) *Fake {
	t.Helper()

	fake := &Fake{Loaded: []string{"model-a"}, Status: http.StatusOK}
	server := httptest.NewServer(fake)
	fake.URL = server.URL

	t.Cleanup(func() {
		server.CloseClientConnections()
		server.Close()
	})

	return fake
}

// Set changes the fake while requests may be running.
func (f *Fake) Set(change func(*Fake)) {
	f.mu.Lock()
	defer f.mu.Unlock()

	change(f)
}

// Bodies are the POST bodies received so far.
func (f *Fake) Bodies() []map[string]any {
	f.mu.Lock()
	defer f.mu.Unlock()

	return append([]map[string]any(nil), f.bodies...)
}

// ServeHTTP answers /api/ps on GET and a chat, probe or unload on POST.
func (f *Fake) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		f.servePs(w)

		return
	}

	var body map[string]any
	if json.NewDecoder(r.Body).Decode(&body) != nil {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	f.mu.Lock()
	f.bodies = append(f.bodies, body)
	unload := body["keep_alive"] == 0.0

	if unload {
		f.Loaded, f.BusyUntil = nil, time.Time{}
	}

	options, _ := body["options"].(map[string]any)
	wait := f.Delay

	if options["num_predict"] == 1.0 {
		wait = time.Until(f.BusyUntil)
	}

	status, raw, gap := f.Status, f.RawBody, f.GenerationDelay
	f.mu.Unlock()

	if unload {
		_, _ = w.Write([]byte(`{"done": true}`))

		return
	}

	if !sleep(r, wait) {
		return
	}

	switch {
	case raw != "":
		w.WriteHeader(status)
		_, _ = w.Write([]byte(raw))
	case body["stream"] != true:
		_, _ = w.Write([]byte(`{"message": {"content": "{\"label\": \"x\"}"}, "done": true, "prompt_eval_count": 7, "eval_count": 3}`))
	default:
		stream(w, r, gap)
	}
}

// servePs lists the loaded models, or answers PsRaw verbatim.
func (f *Fake) servePs(w http.ResponseWriter) {
	f.mu.Lock()
	raw, loaded := f.PsRaw, f.Loaded
	f.mu.Unlock()

	if raw != "" {
		_, _ = w.Write([]byte(raw))

		return
	}

	models := make([]map[string]string, 0, len(loaded))
	for _, name := range loaded {
		models = append(models, map[string]string{"name": name})
	}

	if err := json.NewEncoder(w).Encode(map[string][]map[string]string{"models": models}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// stream answers a chat in two chunks, gap apart.
func stream(w http.ResponseWriter, r *http.Request, gap time.Duration) {
	_, _ = w.Write([]byte(`{"message": {"content": "{\"label\": "}, "done": false}` + "\n"))

	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	if sleep(r, gap) {
		_, _ = w.Write([]byte(`{"message": {"content": "\"x\"}"}, "done": true, "prompt_eval_count": 7, "eval_count": 3}` + "\n"))
	}
}

// sleep waits, and reports false when the client went away first.
func sleep(r *http.Request, wait time.Duration) bool {
	if wait <= 0 {
		return true
	}

	timer := time.NewTimer(wait)
	defer timer.Stop()

	select {
	case <-r.Context().Done():
		return false
	case <-timer.C:
		return true
	}
}
