package host

import (
	"context"
	"io"
	"strconv"
	"strings"
	"time"
)

// LinuxSampler samples a headless Linux server, where host load stands in
// for the user: a server has no keyboard and no battery, and its users are
// the services it hosts.
//
// Idle is the time since the one-minute load average last exceeded maxLoad
// (nil: since the sampler was made), so a busy host blocks idle-only queues
// and releases their attempts exactly as a returning user does. The load
// includes the node's own model, so maxLoad must leave room for it. OnAC is
// always true. Memory counts as warn below minFreePct available.
func LinuxSampler(
	maxLoad *float64,
	minFreePct float64,
	read func(path string) (string, bool),
	now func() time.Time,
	log io.Writer,
) Sampler {
	quietSince := now()

	return func(context.Context) State {
		at := now()
		load := readParsed(read, "/proc/loadavg", parseLoadavg)
		free := readParsed(read, "/proc/meminfo", parseMeminfoFreePct)

		var failed []string
		if load == nil {
			failed = append(failed, "loadavg")
		}

		if free == nil {
			failed = append(failed, "meminfo")
		}

		logFailedReaders(log, failed)

		state := State{OnAC: new(true), Pressure: linuxPressure(free, minFreePct), Load1m: load}
		if load == nil {
			return state
		}

		if maxLoad != nil && *load > *maxLoad {
			quietSince = at
		}

		state.Idle = new(at.Sub(quietSince))

		return state
	}
}

// readParsed is a kernel file's parsed value, or nil when it cannot be read
// or parsed.
func readParsed(read func(string) (string, bool), path string, parse func(string) *float64) *float64 {
	text, ok := read(path)
	if !ok {
		return nil
	}

	return parse(text)
}

// linuxPressure is unknown when unreadable and warn below the free threshold.
func linuxPressure(free *float64, minFreePct float64) Pressure {
	switch {
	case free == nil:
		return PressureUnknown
	case *free < minFreePct:
		return PressureWarn
	default:
		return PressureNormal
	}
}

// parseLoadavg is the first field of /proc/loadavg, the one-minute load.
func parseLoadavg(text string) *float64 {
	fields := strings.Fields(text)
	if len(fields) == 0 {
		return nil
	}

	load, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return nil
	}

	return &load
}

// parseMeminfoFreePct is MemAvailable over MemTotal in percent, or nil when
// either is missing.
func parseMeminfoFreePct(text string) *float64 {
	values := map[string]float64{}

	for line := range strings.SplitSeq(text, "\n") {
		name, rest, _ := strings.Cut(line, ":")
		if name != "MemTotal" && name != "MemAvailable" {
			continue
		}

		parts := strings.Fields(rest)
		if len(parts) == 0 {
			continue
		}

		if kb, err := strconv.ParseUint(parts[0], 10, 64); err == nil {
			values[name] = float64(kb)
		}
	}

	total := values["MemTotal"]
	available, ok := values["MemAvailable"]

	if total == 0 || !ok {
		return nil
	}

	return new(percent * available / total)
}
