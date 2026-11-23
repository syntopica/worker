package host_test

import (
	"bytes"
	"runtime"
	"testing"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/host"
)

func TestDefaultSamplerReadsThePlatformItRunsOn(t *testing.T) {
	t.Parallel()

	node := &config.NodePolicy{MinFreePct: 15}

	state := host.DefaultSampler(node, runtime.GOOS, &bytes.Buffer{})(t.Context())
	if state.Pressure == "" {
		t.Error("the sampler returned no pressure level")
	}
}

func TestDefaultSamplerDegradesOffItsPlatform(t *testing.T) {
	t.Parallel()

	other := "linux"
	if runtime.GOOS == "linux" {
		other = "darwin"
	}

	log := &bytes.Buffer{}

	state := host.DefaultSampler(&config.NodePolicy{MinFreePct: 15}, other, log)(t.Context())
	if state.Pressure != host.PressureUnknown {
		t.Errorf("pressure = %s, want unknown where the readers do not exist", state.Pressure)
	}

	if log.Len() == 0 {
		t.Error("the failed readers were not logged")
	}
}
