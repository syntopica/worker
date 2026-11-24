package tasks

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/syntopica/worker/internal/config"
)

// reapWait bounds the wait for a killed runner to be reaped.
const reapWait = 10 * time.Second

// process is a started runner.
type process struct {
	cmd *exec.Cmd
	// done is closed once the runner has been reaped.
	done chan struct{}
}

// startProcess starts a runner in its own process group, in its workspace,
// with its profile's environment and its output in the two files. Stdin is
// fed from a goroutine, so a large prompt the runner never reads cannot
// block the loop.
func startProcess(profile *config.TaskProfile, call *invocation, workspace string, stdout, stderr *os.File) (*process, error) {
	cmd := exec.Command(call.argv[0], call.argv[1:]...)
	cmd.Dir, cmd.Stdout, cmd.Stderr = workspace, stdout, stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	cmd.Env = slices.DeleteFunc(os.Environ(), func(pair string) bool {
		name, _, _ := strings.Cut(pair, "=")

		return slices.Contains(profile.EnvUnset, name)
	})
	for _, pair := range profile.Env {
		cmd.Env = append(cmd.Env, pair[0]+"="+pair[1])
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("open runner stdin: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start runner: %w", err)
	}

	go func() {
		_, _ = io.WriteString(stdin, call.stdin) // a runner that left is not an error
		_ = stdin.Close()
	}()

	started := &process{cmd: cmd, done: make(chan struct{})}

	go func() {
		_ = cmd.Wait() // the exit status is read from the process state

		close(started.done)
	}()

	return started, nil
}

// supervise waits for the runner while keeping its lease alive and its
// clock bounded. The lease lasts three beats, so it survives two lost ones.
// A fenced, overrunning or stopped runner is killed as a group first:
// stopping the loop must not leave a runner spending quota unattended.
func (p *process) supervise(ctx context.Context, timeout, beat time.Duration, alive func() bool) ending {
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()

	beats := time.NewTicker(beat)
	defer beats.Stop()

	for {
		var how ending

		select {
		case <-p.done:
			return exited
		case <-ctx.Done():
			how = stopped
		case <-deadline.C:
			how = overran
		case <-beats.C:
			if alive() {
				continue
			}

			how = fencedOut
		}

		p.kill()

		return how
	}
}

// exit is the exit status of a runner that ended; -1 when a signal ended it.
func (p *process) exit() int {
	return p.cmd.ProcessState.ExitCode()
}

// kill ends the runner and everything it started, then reaps it; the runner
// itself may already be gone.
func (p *process) kill() {
	_ = syscall.Kill(-p.cmd.Process.Pid, syscall.SIGKILL)

	select {
	case <-p.done:
	case <-time.After(reapWait):
	}
}
