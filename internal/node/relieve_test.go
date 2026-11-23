package node

import (
	"context"
	"testing"
	"time"

	"github.com/syntopica/worker/internal/config"
)

// unloader is a backend that records unloads and answers them with ok.
type unloader struct {
	ok   bool
	sent []string
}

func (*unloader) Resident(context.Context) ([]string, bool) { return nil, true }

func (*unloader) ProbeQuiet(context.Context, *config.ModelPin, time.Duration) bool { return true }

func (u *unloader) Unload(_ context.Context, model string) bool {
	u.sent = append(u.sent, model)

	return u.ok
}

func (*unloader) Restart(context.Context, string) bool { return true }

func (*unloader) Pause(context.Context, time.Duration) {}

func TestBatteryUnloadsOwnedModelsOncePerEpisode(t *testing.T) {
	t.Parallel()

	backend, m := &unloader{}, newMemory()
	m.owned["mine"], m.owned["stuck"], m.failedModel = true, true, "stuck"
	resident := []string{"mine", "other", "stuck"}

	relieve(t.Context(), backend, m, resident, nil, false)
	relieve(t.Context(), backend, m, resident, new(false), false)
	relieve(t.Context(), backend, m, resident, new(false), false)

	// A failed unload keeps its ownership, and AC starts a new episode.
	if len(backend.sent) != 1 || !m.owned["mine"] {
		t.Fatalf("sent %v, owned %v", backend.sent, m.owned)
	}

	backend.ok = true

	relieve(t.Context(), backend, m, resident, new(true), false)
	relieve(t.Context(), backend, m, resident, new(false), false)

	if len(backend.sent) != 2 || m.owned["mine"] || !m.owned["stuck"] {
		t.Fatalf("sent %v, owned %v", backend.sent, m.owned)
	}
}

func TestOwnerActivityUnloadsOwnedModelsOncePerSpell(t *testing.T) {
	t.Parallel()

	backend, m := &unloader{}, newMemory()
	m.owned["mine"] = true

	relieve(t.Context(), backend, m, []string{"mine", "other"}, new(true), true)
	relieve(t.Context(), backend, m, []string{"mine", "other"}, new(true), true)

	if len(backend.sent) != 1 {
		t.Fatalf("sent %v", backend.sent)
	}

	relieve(t.Context(), backend, m, nil, new(true), false)

	if len(m.activeTried) != 0 {
		t.Fatal("an idle spell did not clear the tried models")
	}
}

func TestAFailedDrainIsRetriedEveryFiveMinutesWhileListed(t *testing.T) {
	t.Parallel()

	backend, m := &unloader{}, newMemory()
	m.failedModel = testModel

	for _, seconds := range []int64{100, 399, 400} {
		retryFailedDrain(t.Context(), backend, m, []string{testModel}, at(seconds))
	}

	retryFailedDrain(t.Context(), backend, m, nil, at(900))

	if len(backend.sent) != 2 {
		t.Fatalf("sent %v", backend.sent)
	}
}
