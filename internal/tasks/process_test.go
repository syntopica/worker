package tasks

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/syntopica/worker/internal/config"
)

func TestProcessIsReapedWithUnreadStdin(t *testing.T) {
	t.Parallel()

	for name, want := range map[string]ending{"overrun": overran, "fenced": fencedOut, "stopped": stopped} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assertUnreadProcessReaped(t, want)
		})
	}
}

func assertUnreadProcessReaped(t *testing.T, want ending) {
	t.Helper()

	profile := fake(t, "codex", "exec sleep 60")
	dir := t.TempDir()
	stdout, stderr, err := streams(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _, _ = stdout.Close(), stderr.Close() }()
	const pipePayload = 1 << 20
	call := invocation{argv: []string{profile.Command}, stdin: strings.Repeat("x", pipePayload)}
	runner, err := startProcess(profile, &call, dir, stdout, stderr)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		select {
		case <-runner.done:
		default:
			runner.kill()
		}
	}()

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	timeout := 30 * time.Second
	switch want {
	case overran:
		timeout = 100 * time.Millisecond
	case stopped:
		cancel()
	case fencedOut, exited:
	}
	if got := runner.supervise(ctx, timeout, 20*time.Millisecond, func() bool { return want != fencedOut }); got != want {
		t.Fatalf("runner ending = %v, want %v", got, want)
	}
	assertProcessKilled(t, runner)
}

func assertProcessKilled(t *testing.T, runner *process) {
	t.Helper()

	select {
	case <-runner.done:
		if runner.exit() != -1 {
			t.Fatalf("runner exit = %d, want termination by signal", runner.exit())
		}
	default:
		t.Fatal("runner was not reaped after the timeout")
	}
}

// fake is a profile whose runner executes the test's shell fixture.
func fake(t *testing.T, runner, body string) *config.TaskProfile {
	t.Helper()
	command := filepath.Join(t.TempDir(), "runner")
	if err := os.WriteFile(command, []byte("#!/bin/sh\neval \"$FAKE_BODY\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(command, 0o700); err != nil {
		t.Fatal(err)
	}
	return &config.TaskProfile{
		Name: testProfile, Runner: runner, Model: "m-1", Timeout: 30 * time.Second, Command: command,
		Env: [][2]string{{"FAKE_BODY", body}},
	}
}
