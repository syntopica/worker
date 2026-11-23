package ollama_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/ollama"
)

// scriptedBackend records the drain's steps. The model stays listed until
// goneAfter ("unload", "restart" or never); down makes /api/ps unreachable.
type scriptedBackend struct {
	steps     []string
	listed    bool
	down      bool
	quiet     bool
	goneAfter string
	alive     bool
}

func (b *scriptedBackend) Resident(context.Context) ([]string, bool) {
	if b.down {
		return nil, false
	}

	if b.listed {
		return []string{"model-a", "model-b"}, true
	}

	return []string{"model-b"}, true
}

func (b *scriptedBackend) ProbeQuiet(_ context.Context, _ *config.ModelPin, timeout time.Duration) bool {
	b.steps = append(b.steps, "probe "+timeout.String())

	return b.quiet
}

func (b *scriptedBackend) Unload(_ context.Context, model string) bool {
	b.steps = append(b.steps, "unload "+model)
	b.listed = b.listed && b.goneAfter != "unload"

	return true
}

func (b *scriptedBackend) Restart(_ context.Context, label string) bool {
	b.steps = append(b.steps, "restart "+label)
	b.listed = b.listed && b.goneAfter != "restart"

	return true
}

func (b *scriptedBackend) Pause(_ context.Context, wait time.Duration) {
	b.steps = append(b.steps, "pause "+wait.String())
}

// drain runs a drain and returns its verdict and steps.
func drain(t *testing.T, backend *scriptedBackend) (quiet bool, steps []string) {
	t.Helper()

	quiet = ollama.Drain(t.Context(), backend, pin(), "svc.label", func(context.Context) bool {
		backend.steps = append(backend.steps, "beat")

		return backend.alive
	})

	return quiet, backend.steps
}

func repeat(times int, steps ...string) []string {
	out := make([]string, 0, len(steps)*times)
	for range times {
		out = append(out, steps...)
	}

	return out
}

func TestDrain(t *testing.T) {
	t.Parallel()

	polls := repeat(5, "beat", "pause 2s")
	cases := []struct {
		name    string
		backend scriptedBackend
		quiet   bool
		steps   []string
	}{
		{"an unlisted model is quiet", scriptedBackend{alive: true}, true, []string{"beat"}},
		{"a probe that answers is quiet", scriptedBackend{alive: true, listed: true, quiet: true}, true, []string{"beat", "probe 20s"}},
		{
			"an unload follows three failed probes",
			scriptedBackend{alive: true, listed: true, goneAfter: "unload"},
			true,
			append(repeat(3, "beat", "probe 20s"), "beat", "unload model-a", "beat"),
		},
		{
			"a restart follows an unload that left the model resident",
			scriptedBackend{alive: true, listed: true, goneAfter: "restart"},
			true,
			append(append(append(repeat(3, "beat", "probe 20s"), "beat", "unload model-a"), polls...), "beat", "restart svc.label", "beat"),
		},
		{
			"a model that survives the restart is a drain failure",
			scriptedBackend{alive: true, listed: true},
			false,
			append(append(append(append(repeat(3, "beat", "probe 20s"), "beat", "unload model-a"), polls...), "beat", "restart svc.label"), polls...),
		},
		{"a fenced drain never unloads or restarts", scriptedBackend{listed: true}, false, []string{"beat", "probe 20s"}},
		{"a fenced drain still sees a quiet backend", scriptedBackend{listed: true, quiet: true}, true, []string{"beat", "probe 20s"}},
		{
			"an unreachable listing never restarts",
			scriptedBackend{alive: true, down: true},
			false,
			append(append(repeat(3, "beat", "pause 2s"), "beat", "unload model-a"), polls...),
		},
	}

	for _, c := range cases {
		quiet, steps := drain(t, &c.backend)
		if quiet != c.quiet {
			t.Errorf("%s: quiet = %v, want %v", c.name, quiet, c.quiet)
		}

		if !reflect.DeepEqual(steps, c.steps) {
			t.Errorf("%s: steps = %v\nwant %v", c.name, steps, c.steps)
		}
	}
}
