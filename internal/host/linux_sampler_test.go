package host_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/syntopica/worker/internal/host"
)

// procFiles serves /proc from a map the test mutates; a missing path fails.
func procFiles(files map[string]string) func(string) (string, bool) {
	return func(path string) (string, bool) {
		text, ok := files[path]

		return text, ok
	}
}

func TestLinuxSamplerLoadStandsInForTheUser(t *testing.T) {
	t.Parallel()

	files := map[string]string{
		"/proc/loadavg": "0.50 0.40 0.30 1/200 999\n",
		"/proc/meminfo": "MemTotal: 1000 kB\nMemAvailable: 400 kB\n",
	}
	at := time.Unix(1000, 0)
	log := &bytes.Buffer{}
	sample := host.LinuxSampler(new(2.0), 15, procFiles(files), func() time.Time { return at }, log)

	at = at.Add(90 * time.Second)

	state := sample(t.Context())
	if *state.Idle != 90*time.Second {
		t.Errorf("idle = %v, want 90s since the sampler was made", *state.Idle)
	}

	if !*state.OnAC || state.Pressure != host.PressureNormal || *state.Load1m != 0.5 {
		t.Errorf("state = %+v, want AC, normal pressure and load 0.5", state)
	}

	files["/proc/loadavg"] = "2.50 1 1 1/1 1\n"
	at = at.Add(10 * time.Second)

	if state = sample(t.Context()); *state.Idle != 0 {
		t.Errorf("idle = %v, want 0 on a busy host", *state.Idle)
	}

	files["/proc/loadavg"] = "1.00 1 1 1/1 1\n"
	at = at.Add(20 * time.Second)

	if state = sample(t.Context()); *state.Idle != 20*time.Second {
		t.Errorf("idle = %v, want 20s since the load dropped", *state.Idle)
	}

	if log.Len() != 0 {
		t.Errorf("unexpected log: %q", log.String())
	}
}

func TestLinuxSamplerWithoutALimitNeverCountsAsBusy(t *testing.T) {
	t.Parallel()

	files := map[string]string{"/proc/loadavg": "64.0 1 1\n", "/proc/meminfo": "MemTotal: 1000 kB\nMemAvailable: 100 kB\n"}
	at := time.Unix(1000, 0)
	sample := host.LinuxSampler(nil, 15, procFiles(files), func() time.Time { return at }, &bytes.Buffer{})

	at = at.Add(time.Minute)

	state := sample(t.Context())
	if *state.Idle != time.Minute {
		t.Errorf("idle = %v, want one minute", *state.Idle)
	}

	if state.Pressure != host.PressureWarn {
		t.Errorf("pressure = %s, want warn below the free threshold", state.Pressure)
	}
}

func TestLinuxSamplerNamesTheReadersThatFailed(t *testing.T) {
	t.Parallel()

	cases := map[string]map[string]string{
		"missing":    {},
		"unparsable": {"/proc/loadavg": "busy\n", "/proc/meminfo": "MemTotal: 0 kB\nMemAvailable: 5 kB\n"},
		"empty":      {"/proc/loadavg": "", "/proc/meminfo": "MemTotal: 1000 kB\nMemAvailable:\n"},
	}

	for name, files := range cases {
		log := &bytes.Buffer{}

		state := host.LinuxSampler(nil, 15, procFiles(files), time.Now, log)(t.Context())
		if state.Idle != nil || state.Load1m != nil || state.Pressure != host.PressureUnknown {
			t.Errorf("%s: state = %+v, want unreadable", name, state)
		}

		if got := log.String(); got != "worker: host readers failed: loadavg,meminfo\n" {
			t.Errorf("%s: log = %q", name, got)
		}
	}
}
