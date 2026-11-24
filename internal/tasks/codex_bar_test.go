package tasks

import "testing"

func TestCodexBarReadsTheFirstReportsUsage(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if got := (CodexBar{Command: dir + "/absent"}).Usage(t.Context(), "codex"); got != nil {
		t.Errorf("a missing CodexBar read as %v", got)
	}

	if got := (CodexBar{Command: "true"}).Usage(t.Context(), "codex"); got != nil {
		t.Errorf("an empty report read as %v", got)
	}

	if got := (CodexBar{Command: "echo"}).Usage(t.Context(), "codex"); got != nil {
		t.Errorf("a report that is not JSON read as %v", got)
	}
}
