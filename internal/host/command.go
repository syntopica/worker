package host

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

// Command returns a runner of read-only system commands: stdout on exit 0,
// else false (spec 7: failure = busy).
//
// A failure is logged as the command's base name and the kind of failure
// (timeout, exit status or OS error name) and nothing else, so a flapping
// host reader can be told apart from a slow one.
func Command(timeout time.Duration, log io.Writer) func(ctx context.Context, name string, args ...string) (string, bool) {
	return func(ctx context.Context, name string, args ...string) (string, bool) {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		out, err := exec.CommandContext(ctx, name, args...).Output()
		if err == nil {
			return string(out), true
		}

		_, _ = fmt.Fprintf(log, "worker: command failed: %s %s\n", filepath.Base(name), failureKind(ctx, err, timeout))

		return "", false
	}
}

// failureKind names why a command failed without quoting anything it wrote.
func failureKind(ctx context.Context, err error, timeout time.Duration) string {
	var exit *exec.ExitError

	switch {
	case errors.Is(ctx.Err(), context.DeadlineExceeded):
		return "timeout " + strconv.FormatFloat(timeout.Seconds(), 'g', -1, 64) + "s"
	case errors.As(err, &exit) && exit.ExitCode() >= 0:
		return "exit " + strconv.Itoa(exit.ExitCode())
	case errors.Is(err, fs.ErrNotExist):
		// shortcut: only the two errno names seen in practice, add more when one shows as unknown.
		return "oserror ENOENT"
	case errors.Is(err, fs.ErrPermission):
		return "oserror EACCES"
	default:
		return "oserror unknown"
	}
}
