package host

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/syntopica/worker/internal/config"
)

const readerTimeout = 5 * time.Second

// DefaultSampler chooses the host sampler for the platform a node runs on:
// Linux reads load and memory from /proc; anything else is the macOS reader.
func DefaultSampler(node *config.NodePolicy, goos string, log io.Writer) Sampler {
	if strings.HasPrefix(goos, "linux") {
		return LinuxSampler(node.MaxLoad, node.MinFreePct, readTextFile, time.Now, log)
	}

	sample := DarwinSampler(Command(readerTimeout, log), node.MinFreePct, log)

	return HoldPowerSource(BridgeUnreadable(sample, time.Now), time.Now)
}

// readTextFile reads a small kernel file; false when it cannot be read.
func readTextFile(path string) (string, bool) {
	raw, err := os.ReadFile(path)

	return string(raw), err == nil
}
