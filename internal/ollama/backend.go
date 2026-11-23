package ollama

import (
	"context"
	"time"

	"github.com/syntopica/worker/internal/config"
)

// Backend is what a drain needs from the model server.
type Backend interface {
	// Resident lists the models in memory; false when the server does not
	// answer.
	Resident(ctx context.Context) ([]string, bool)
	// ProbeQuiet reports whether a one-token request got its turn in time.
	ProbeQuiet(ctx context.Context, pin *config.ModelPin, timeout time.Duration) bool
	// Unload asks the server to drop a model from memory now.
	Unload(ctx context.Context, model string) bool
	// Restart restarts the model server's LaunchAgent.
	Restart(ctx context.Context, label string) bool
	// Pause waits between drain steps.
	Pause(ctx context.Context, wait time.Duration)
}
