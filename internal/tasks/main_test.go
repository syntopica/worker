package tasks

import (
	"os"
	"path/filepath"
	"testing"
)

// fakeRunner is a runner whose behaviour is the shell text in FAKE_BODY.
// It is written once, before any test starts a process: an executable still
// open for writing in one test cannot be started by another.
var fakeRunner string //nolint:gochecknoglobals // set once, before any test runs.

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "worker-fake-runner-")
	if err != nil {
		panic(err)
	}

	fakeRunner = filepath.Join(dir, "runner")

	//nolint:gosec // G306: a test script has to be executable.
	if err = os.WriteFile(fakeRunner, []byte("#!/bin/sh\neval \"$FAKE_BODY\"\n"), 0o700); err != nil {
		panic(err)
	}

	code := m.Run()

	_ = os.RemoveAll(dir)

	os.Exit(code)
}
