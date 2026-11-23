package host_test

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/syntopica/worker/internal/host"
)

const (
	idleKey     = "ioreg -c IOHIDSystem -d 4"
	batteryKey  = "ioreg -r -c AppleSmartBattery -d 1"
	pmsetKey    = "pmset -g ps"
	pressureKey = "sysctl -n kern.memorystatus_vm_pressure_level"
	freeKey     = "sysctl -n kern.memorystatus_level"
)

// healthy answers every reader: 42 s idle, on AC, normal pressure, 60% free.
func healthy() map[string]string {
	return map[string]string{
		idleKey:     `    | "HIDIdleTime" = 42000000000` + "\n",
		batteryKey:  `      "ExternalConnected" = Yes` + "\n",
		pmsetKey:    "Now drawing from 'AC Power'\n -InternalBattery-0\t100%\n",
		pressureKey: "1\n",
		freeKey:     "60\n",
	}
}

// commands answers each command line from a map and records what ran; a
// missing key is a failed command.
func commands(answers map[string]string, ran *[]string) func(context.Context, string, ...string) (string, bool) {
	return func(_ context.Context, name string, args ...string) (string, bool) {
		key := name[strings.LastIndex(name, "/")+1:] + " " + strings.Join(args, " ")
		*ran = append(*ran, key)
		out, ok := answers[key]

		return out, ok
	}
}

// sampleDarwin samples once and returns the state, the log and what ran.
func sampleDarwin(t *testing.T, answers map[string]string) (state host.State, logged string, ran []string) {
	t.Helper()

	log := &bytes.Buffer{}
	state = host.DarwinSampler(commands(answers, &ran), 15, log)(t.Context())

	return state, log.String(), ran
}

func TestDarwinSamplerReadsAHealthyHost(t *testing.T) {
	t.Parallel()

	state, log, ran := sampleDarwin(t, healthy())
	if state.Idle == nil || *state.Idle != 42*time.Second {
		t.Errorf("idle = %v, want 42s", state.Idle)
	}

	if state.OnAC == nil || !*state.OnAC || state.Pressure != host.PressureNormal || state.Load1m != nil {
		t.Errorf("state = %+v", state)
	}

	if log != "" {
		t.Errorf("unexpected log: %q", log)
	}

	if strings.Contains(strings.Join(ran, ";"), "pmset") {
		t.Error("pmset ran although the battery flag answered")
	}
}

func TestDarwinSamplerPowerSource(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		battery, ps   string
		want          *bool
		log           string
		pmsetAttempts int
	}{
		{"battery flag says no", `"ExternalConnected" = No`, "", new(false), "", 0},
		{"desktop falls back to pmset", "no battery here", "Now drawing from 'AC Power'\n", new(true), "", 1},
		{"pmset on battery", "", "Now drawing from 'Battery Power'\n", new(false), "", 1},
		{"pmset unparsable", "", "Now drawing from 'UPS'\n", nil, "worker: host readers failed: pmset_parse\n", 1},
		{"pmset fails twice", "", "", nil, "worker: host readers failed: pmset\n", 2},
	}

	for _, c := range cases {
		answers := healthy()
		delete(answers, batteryKey)
		delete(answers, pmsetKey)

		if c.battery != "" {
			answers[batteryKey] = c.battery
		}

		if c.ps != "" {
			answers[pmsetKey] = c.ps
		}

		state, log, ran := sampleDarwin(t, answers)

		switch {
		case (state.OnAC == nil) != (c.want == nil):
			t.Errorf("%s: on_ac = %v", c.name, state.OnAC)
		case c.want != nil && *state.OnAC != *c.want:
			t.Errorf("%s: on_ac = %v, want %v", c.name, *state.OnAC, *c.want)
		}

		if log != c.log {
			t.Errorf("%s: log = %q, want %q", c.name, log, c.log)
		}

		if got := strings.Count(strings.Join(ran, ";"), "pmset"); got != c.pmsetAttempts {
			t.Errorf("%s: pmset ran %d times, want %d", c.name, got, c.pmsetAttempts)
		}
	}
}

func TestDarwinSamplerPressure(t *testing.T) {
	t.Parallel()

	cases := []struct {
		level, free string
		want        host.Pressure
	}{
		{"1", "60", host.PressureNormal},
		{"2", "33", host.PressureNormal},
		{"2", "14.9", host.PressureWarn},
		{"4", "80", host.PressureCritical},
		{"4", "", host.PressureCritical},
		{"1", "", host.PressureUnknown},
		{"1", "101", host.PressureUnknown},
		{"1", "-1", host.PressureUnknown},
		{"1", "NaN", host.PressureUnknown},
		{"3", "60", host.PressureUnknown},
	}

	for _, c := range cases {
		answers := healthy()
		answers[pressureKey] = c.level + "\n"
		delete(answers, freeKey)

		if c.free != "" {
			answers[freeKey] = c.free + "\n"
		}

		if state, _, _ := sampleDarwin(t, answers); state.Pressure != c.want {
			t.Errorf("level %s free %q: pressure = %s, want %s", c.level, c.free, state.Pressure, c.want)
		}
	}
}

func TestDarwinSamplerNamesTheReadersThatFailed(t *testing.T) {
	t.Parallel()

	state, log, _ := sampleDarwin(t, map[string]string{})
	if state.Idle != nil || state.OnAC != nil || state.Pressure != host.PressureUnknown {
		t.Errorf("state = %+v, want unreadable", state)
	}

	if log != "worker: host readers failed: ioreg,pmset,pressure_level\n" {
		t.Errorf("log = %q", log)
	}

	answers := healthy()
	answers[idleKey] = "no idle time here"
	answers[pressureKey] = "seven"

	_, log, _ = sampleDarwin(t, answers)
	if log != "worker: host readers failed: ioreg_parse,pressure_level_parse\n" {
		t.Errorf("log = %q", log)
	}

	answers[idleKey] = `"HIDIdleTime" = 99999999999999999999999`

	if state, _, _ = sampleDarwin(t, answers); state.Idle != nil {
		t.Error("an idle time that overflows was accepted")
	}
}
