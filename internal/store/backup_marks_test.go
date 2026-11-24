package store

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func payloadDir(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()

	err := os.WriteFile(filepath.Join(dir, payloadFile), []byte("x"), privateFile)
	if err != nil {
		t.Fatal(err)
	}

	return dir
}

func TestAFailingExclusionLogsOneLineAndIsRetried(t *testing.T) {
	t.Parallel()

	dir := payloadDir(t)
	marks := newBackupMarks()
	calls := 0
	lines := []string{}
	marks.log = func(line string) { lines = append(lines, line) }
	marks.run = func(context.Context, string) error {
		calls++
		if calls == 1 {
			return &fs.PathError{Op: "exec", Path: "no xattr here", Err: fs.ErrNotExist}
		}

		return nil
	}

	for range 3 {
		marks.exclude(t.Context(), dir)
	}

	if calls != 2 {
		t.Fatalf("calls = %d, want a retry after the failure and none after the success", calls)
	}

	if len(lines) != 1 || lines[0] != "worker: backup exclusion failed: *fs.PathError" {
		t.Fatalf("lines = %q", lines)
	}
}

func TestARecreatedFileIsMarkedAgain(t *testing.T) {
	t.Parallel()

	dir := payloadDir(t)
	marks := newBackupMarks()
	calls := 0
	marks.run = func(context.Context, string) error {
		calls++

		return nil
	}
	marks.exclude(t.Context(), dir)

	path := filepath.Join(dir, payloadFile)

	err := os.WriteFile(path+".new", []byte("y"), privateFile)
	if err == nil {
		err = os.Rename(path+".new", path)
	}

	if err != nil {
		t.Fatal(err)
	}

	marks.exclude(t.Context(), dir)

	if calls != 2 {
		t.Fatalf("calls = %d", calls)
	}
}

func TestAnExitStatusIsNamedWithoutItsOutput(t *testing.T) {
	t.Parallel()

	err := exec.CommandContext(t.Context(), "/bin/sh", "-c", "exit 3").Run()
	if got := errorClass(err); got != "exit 3" {
		t.Fatalf("class = %q", got)
	}

	if got := errorClass(errors.New("secret path")); strings.Contains(got, "secret") {
		t.Fatalf("class = %q", got)
	}
}

func TestPayloadFilesAreExcludedFromTimeMachine(t *testing.T) {
	t.Parallel()

	if runtime.GOOS != "darwin" {
		t.Skip("Time Machine exists on macOS only")
	}

	dir := t.TempDir()
	freshStore(t, dir)

	attribute := func(name string) string {
		//nolint:gosec // fixed binary; the path is the test's own directory
		out, _ := exec.CommandContext(t.Context(), "/usr/bin/xattr", "-p", backupAttribute, filepath.Join(dir, name)).
			CombinedOutput()

		return string(out)
	}

	if got := attribute(payloadFile); !strings.Contains(got, "com.apple.backupd") {
		t.Fatalf("payload attribute = %q", got)
	}

	if got := attribute(metaFile); strings.Contains(got, "com.apple.backupd") {
		t.Fatalf("the metadata file is excluded: %q", got)
	}
}
