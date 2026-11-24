package node

import (
	"slices"
	"testing"

	"github.com/syntopica/worker/internal/config"
)

func TestHostReportKeepsOnlyConfiguredModelNames(t *testing.T) {
	t.Parallel()

	node := &Node{}
	cfg := &config.Config{Models: map[string]config.ModelPin{"model": {}}}
	state := idleState()
	report := node.hostReport(cfg, state, []string{"model", "other-model"}, false, "", at(0))
	if !slices.Equal(report.Resident, []string{"model", "unconfigured"}) {
		t.Errorf("resident = %v, want managed names and a fixed marker", report.Resident)
	}
	if !slices.Equal(report.Unexpected, []string{"unconfigured"}) {
		t.Errorf("unexpected = %v, want one fixed marker", report.Unexpected)
	}
}
