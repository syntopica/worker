package ollama

import (
	"context"
	"slices"
	"time"

	"github.com/syntopica/worker/internal/config"
)

const (
	probeRounds  = 3
	probeTimeout = 20 * time.Second
	drainPause   = 2 * time.Second
	unlistPolls  = 5
)

// Drain waits until the backend is quiet: probe, then unload, then restart
// (spec 7). True once it is confirmed quiet; false is a drain failure.
//
//  1. Up to three one-token probes, each only while the model is resident (a
//     probe of an absent model would load it). A model /api/ps no longer
//     lists is quiet. A fenced heartbeat stops here: never unload or restart.
//  2. Unload; quiet once /api/ps stops listing the model. No probe follows.
//  3. Restart only when the model is still resident after the unload, and
//     therefore still busy: every probe timed out. An unreachable /api/ps
//     never triggers a restart.
//
// A heartbeat runs between every step, so a long drain keeps its lease.
func Drain(
	ctx context.Context,
	backend Backend,
	pin *config.ModelPin,
	label string,
	heartbeat func(context.Context) bool,
) bool {
	if quiet, settled := probeUntilQuiet(ctx, backend, pin, heartbeat); settled {
		return quiet
	}

	heartbeat(ctx)
	backend.Unload(ctx, pin.Name)

	if answered, gone := waitUnlisted(ctx, backend, pin.Name, heartbeat); !answered || gone {
		return answered
	}

	heartbeat(ctx)
	backend.Restart(ctx, label)

	answered, gone := waitUnlisted(ctx, backend, pin.Name, heartbeat)

	return answered && gone
}

// waitUnlisted polls /api/ps until a model is no longer listed, heartbeating
// between polls. answered is whether the last poll got a listing; gone is
// whether that listing lacks the model.
func waitUnlisted(
	ctx context.Context,
	backend Backend,
	model string,
	heartbeat func(context.Context) bool,
) (answered, gone bool) {
	for range unlistPolls {
		heartbeat(ctx)

		listed, ok := backend.Resident(ctx)
		answered = ok

		if ok && !slices.Contains(listed, model) {
			return true, true
		}

		backend.Pause(ctx, drainPause)
	}

	return answered, false
}

// probeUntilQuiet is the first drain step. settled is false when every probe
// timed out with the lease still held, which is when the unload follows.
func probeUntilQuiet(
	ctx context.Context,
	backend Backend,
	pin *config.ModelPin,
	heartbeat func(context.Context) bool,
) (quiet, settled bool) {
	for range probeRounds {
		fenced := !heartbeat(ctx)
		listed, ok := backend.Resident(ctx)

		if ok && !slices.Contains(listed, pin.Name) {
			return true, true
		}

		if ok && backend.ProbeQuiet(ctx, pin, probeTimeout) {
			return true, true
		}

		if fenced {
			return false, true
		}

		if !ok {
			backend.Pause(ctx, drainPause)
		}
	}

	return false, false
}
