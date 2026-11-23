package host_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/syntopica/worker/internal/host"
)

func TestCommandReturnsStdoutOnSuccess(t *testing.T) {
	t.Parallel()

	log := &bytes.Buffer{}

	out, ok := host.Command(5*time.Second, log)(t.Context(), "/bin/echo", "ready")
	if !ok || out != "ready\n" {
		t.Errorf("out = %q, ok = %v", out, ok)
	}

	if log.Len() != 0 {
		t.Errorf("unexpected log: %q", log.String())
	}
}

func TestCommandLogsOnlyTheKindOfFailure(t *testing.T) {
	t.Parallel()

	locked := filepath.Join(t.TempDir(), "locked")
	if err := os.WriteFile(locked, []byte("#!/bin/sh\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name    string
		timeout time.Duration
		argv    []string
		want    string
	}{
		{"exit status", 5 * time.Second, []string{"/bin/sh", "-c", "echo private >&2; exit 3"}, "worker: command failed: sh exit 3\n"},
		{"timeout", 100 * time.Millisecond, []string{"/bin/sleep", "5"}, "worker: command failed: sleep timeout 0.1s\n"},
		{"missing", 5 * time.Second, []string{"/nonexistent/reader"}, "worker: command failed: reader oserror ENOENT\n"},
		{"not executable", 5 * time.Second, []string{locked}, "worker: command failed: locked oserror EACCES\n"},
	}

	for _, c := range cases {
		log := &bytes.Buffer{}

		out, ok := host.Command(c.timeout, log)(t.Context(), c.argv[0], c.argv[1:]...)
		if ok || out != "" {
			t.Errorf("%s: out = %q, ok = %v", c.name, out, ok)
		}

		if log.String() != c.want {
			t.Errorf("%s: log = %q, want %q", c.name, log.String(), c.want)
		}
	}
}
