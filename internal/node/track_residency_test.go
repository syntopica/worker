package node

import (
	"bytes"
	"testing"

	"github.com/syntopica/worker/internal/config"
)

func TestResidencyChangesAreCountedForAnHourAndLogged(t *testing.T) {
	t.Parallel()

	m, log := newMemory(), &bytes.Buffer{}
	cfg := &config.Config{Models: map[string]config.ModelPin{"a": {}, "b": {}}}

	if loads, unloads := trackResidency(cfg, m, []string{"a"}, at(0), log); loads+unloads != 0 || log.Len() != 0 {
		t.Fatal("the first sample counted a change")
	}

	trackResidency(cfg, m, []string{"b", "c"}, at(60), log)

	loads, unloads := trackResidency(cfg, m, []string{"b"}, at(120), log)
	if loads != 2 || unloads != 2 {
		t.Fatalf("loads %d, unloads %d", loads, unloads)
	}

	want := "worker: residency load model=b at=1970-01-01T00:01:00Z\n" +
		"worker: residency load model=unconfigured at=1970-01-01T00:01:00Z\n" +
		"worker: residency unload model=a at=1970-01-01T00:01:00Z\n" +
		"worker: residency unload model=unconfigured at=1970-01-01T00:02:00Z\n"
	if log.String() != want {
		t.Fatalf("log: %q", log.String())
	}

	if loads, unloads = trackResidency(cfg, m, []string{"b"}, at(3700), log); loads != 0 || unloads != 1 {
		t.Fatalf("after an hour: loads %d, unloads %d", loads, unloads)
	}
}
