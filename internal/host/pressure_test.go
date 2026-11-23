package host_test

import (
	"testing"

	"github.com/syntopica/worker/internal/host"
)

func TestPressurePressed(t *testing.T) {
	t.Parallel()

	want := map[host.Pressure]bool{
		host.PressureNormal:   false,
		host.PressureWarn:     true,
		host.PressureCritical: true,
		host.PressureUnknown:  false,
	}

	for level, pressed := range want {
		if level.Pressed() != pressed {
			t.Errorf("%s.Pressed() = %v, want %v", level, level.Pressed(), pressed)
		}
	}
}
