package ollama_test

import (
	"context"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/syntopica/worker/internal/ollama"
	"github.com/syntopica/worker/internal/ollamatest"
)

// server is the client of a fake.
func server(fake *ollamatest.Fake) *ollama.Server {
	return &ollama.Server{URL: fake.URL, HTTP: http.DefaultClient}
}

func TestServerResidentAndUnload(t *testing.T) {
	t.Parallel()

	fake := ollamatest.New(t)
	backend := server(fake)

	if names, ok := backend.Resident(t.Context()); !ok || !reflect.DeepEqual(names, []string{"model-a"}) {
		t.Errorf("resident = %v, %v", names, ok)
	}

	if !backend.Unload(t.Context(), "model-a") {
		t.Error("unload failed")
	}

	if names, ok := backend.Resident(t.Context()); !ok || len(names) != 0 {
		t.Errorf("resident after unload = %v, %v", names, ok)
	}

	want := map[string]any{"model": "model-a", "keep_alive": 0.0}
	if got := fake.Bodies(); len(got) != 1 || !reflect.DeepEqual(got[0], want) {
		t.Errorf("unload body = %v", got)
	}
}

func TestServerResidentRefusesAnOddListing(t *testing.T) {
	t.Parallel()

	for _, raw := range []string{`{"models":[{}]}`, `{"models":5}`, `[1]`, `not json`} {
		fake := ollamatest.New(t)
		fake.PsRaw = raw

		if names, ok := server(fake).Resident(t.Context()); ok {
			t.Errorf("%s: resident = %v, want no answer", raw, names)
		}
	}

	down := &ollama.Server{URL: "http://127.0.0.1:1", HTTP: http.DefaultClient}
	if _, ok := down.Resident(t.Context()); ok || down.Unload(t.Context(), "model-a") {
		t.Error("an unreachable server answered")
	}

	if _, ok := (&ollama.Server{URL: "://bad", HTTP: http.DefaultClient}).Resident(t.Context()); ok {
		t.Error("a bad URL answered")
	}
}

func TestServerProbeReportsQuietOnlyAfterTheBackendFrees(t *testing.T) {
	t.Parallel()

	fake := ollamatest.New(t)
	fake.BusyUntil = time.Now().Add(400 * time.Millisecond)
	backend := server(fake)

	if backend.ProbeQuiet(t.Context(), pin(), 100*time.Millisecond) {
		t.Error("a busy backend probed quiet")
	}

	if !backend.ProbeQuiet(t.Context(), pin(), 5*time.Second) {
		t.Error("the backend never probed quiet")
	}

	want := map[string]any{
		"model":      "model-a",
		"messages":   []any{map[string]any{"role": "user", "content": "."}},
		"stream":     false,
		"think":      false,
		"keep_alive": "30m",
		"options":    map[string]any{"num_ctx": 40960.0, "num_predict": 1.0},
	}
	if got := fake.Bodies()[0]; !reflect.DeepEqual(got, want) {
		t.Errorf("probe body = %v", got)
	}

	fake.Set(func(f *ollamatest.Fake) { f.Status, f.RawBody = http.StatusInternalServerError, "boom" })

	if backend.ProbeQuiet(t.Context(), pin(), 5*time.Second) {
		t.Error("a failing backend probed quiet")
	}
}

func TestServerRestartUsesTheUsersGUIDomain(t *testing.T) {
	t.Parallel()

	var seen []string

	backend := &ollama.Server{Run: func(_ context.Context, name string, args ...string) (string, bool) {
		seen = append([]string{name}, args...)

		return "", len(args) == 3
	}}

	if !backend.Restart(t.Context(), "svc.label") {
		t.Error("restart failed")
	}

	want := []string{"/bin/launchctl", "kickstart", "-k", "gui/" + strconv.Itoa(os.Getuid()) + "/svc.label"}
	if !reflect.DeepEqual(seen, want) {
		t.Errorf("ran %v, want %v", seen, want)
	}
}

func TestServerPauseEndsWithTheContext(t *testing.T) {
	t.Parallel()

	backend := &ollama.Server{}
	backend.Pause(t.Context(), time.Millisecond)

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	began := time.Now()
	backend.Pause(ctx, time.Minute)

	if time.Since(began) > 5*time.Second {
		t.Error("the pause outlived its context")
	}
}
