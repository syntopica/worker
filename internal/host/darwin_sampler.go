package host

import (
	"context"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	ioregPath  = "/usr/sbin/ioreg"
	pmsetPath  = "/usr/bin/pmset"
	sysctlPath = "/usr/sbin/sysctl"
)

// DarwinSampler reads idle time, power source and memory pressure on macOS
// with five short commands at most; each failure degrades to an unreadable
// field.
//
// A failed reader is logged by its fixed name only, to tell which one makes
// host_state_unreadable flap.
func DarwinSampler(
	run func(ctx context.Context, name string, args ...string) (string, bool),
	minFreePct float64,
	log io.Writer,
) Sampler {
	return func(ctx context.Context) State {
		var failed []string

		idle, name := readIdle(ctx, run)
		failed = appendFailed(failed, name)

		onAC, name := readOnAC(ctx, run)
		failed = appendFailed(failed, name)

		level := PressureUnknown

		raw, ok := run(ctx, sysctlPath, "-n", "kern.memorystatus_vm_pressure_level")
		if ok {
			level = parsePressure(raw)
		}

		switch {
		case !ok:
			failed = append(failed, "pressure_level")
		case level == PressureUnknown:
			failed = append(failed, "pressure_level_parse")
		}

		var free *float64
		if raw, ok = run(ctx, sysctlPath, "-n", "kern.memorystatus_level"); ok {
			free = parseFreePct(raw)
		}

		logFailedReaders(log, failed)

		return State{Idle: idle, OnAC: onAC, Pressure: effectivePressure(level, free, minFreePct)}
	}
}

// appendFailed adds a failed reader's name, when there is one.
func appendFailed(failed []string, name string) []string {
	if name == "" {
		return failed
	}

	return append(failed, name)
}

// readIdle is the time since the last input, or the name of what failed.
func readIdle(ctx context.Context, run func(context.Context, string, ...string) (string, bool)) (idle *time.Duration, failed string) {
	raw, ok := run(ctx, ioregPath, "-c", "IOHIDSystem", "-d", "4")
	if !ok {
		return nil, "ioreg"
	}

	// HIDIdleTime is in nanoseconds and covers every user at the console.
	match := regexp.MustCompile(`"HIDIdleTime" = (\d+)`).FindStringSubmatch(raw)
	if match == nil {
		return nil, "ioreg_parse"
	}

	nanos, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil {
		return nil, "ioreg_parse"
	}

	return new(time.Duration(nanos)), ""
}

// readOnAC is whether the machine is on AC power, or the name of what failed.
//
// The battery's own flag answers in ~10 ms; pmset waits on powerd and timed
// out 3,426 times by 2026-10-17, so it is only the desktop fallback.
func readOnAC(ctx context.Context, run func(context.Context, string, ...string) (string, bool)) (onAC *bool, failed string) {
	if raw, ok := run(ctx, ioregPath, "-r", "-c", "AppleSmartBattery", "-d", "1"); ok {
		// No match when ioreg lists no battery (a desktop) or the flag is absent.
		match := regexp.MustCompile(`"ExternalConnected"\s*=\s*(Yes|No)`).FindStringSubmatch(raw)
		if match != nil {
			return new(match[1] == "Yes"), ""
		}
	}

	raw, ok := run(ctx, pmsetPath, "-g", "ps")
	if !ok {
		// pmset alone was seen failing transiently (2026-09-30), which blocked
		// admission for a whole sample; one immediate retry absorbs it.
		raw, ok = run(ctx, pmsetPath, "-g", "ps")
	}

	if !ok {
		return nil, "pmset"
	}

	first, _, _ := strings.Cut(raw, "\n")

	switch {
	case strings.Contains(first, "AC Power"):
		return new(true), ""
	case strings.Contains(first, "Battery Power"):
		return new(false), ""
	default:
		return nil, "pmset_parse"
	}
}

// parsePressure maps kern.memorystatus_vm_pressure_level to a named level;
// anything unexpected is unknown, which blocks work.
func parsePressure(text string) Pressure {
	switch strings.TrimSpace(text) {
	case "1":
		return PressureNormal
	case "2":
		return PressureWarn
	case "4":
		return PressureCritical
	default:
		return PressureUnknown
	}
}

// parseFreePct parses kern.memorystatus_level, the free-memory percentage;
// nil for anything that is not a percentage, which blocks work.
func parseFreePct(text string) *float64 {
	value, err := strconv.ParseFloat(strings.TrimSpace(text), 64)
	if err != nil || !(value >= 0 && value <= percent) {
		return nil
	}

	return &value
}

// effectivePressure combines the kernel's level with free memory (spec
// amendment 2026-09-30): warn counts only below minFreePct, because the
// kernel raises it with a third free. Critical always counts, and an
// unreadable free percentage is unknown.
func effectivePressure(level Pressure, free *float64, minFreePct float64) Pressure {
	switch {
	case level == PressureCritical:
		return level
	case level == PressureUnknown || free == nil:
		return PressureUnknown
	case level == PressureWarn && *free < minFreePct:
		return PressureWarn
	default:
		return PressureNormal
	}
}
