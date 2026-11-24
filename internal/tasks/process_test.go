package tasks

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/syntopica/worker/internal/config"
)

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
