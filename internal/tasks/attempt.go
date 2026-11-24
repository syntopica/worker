package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/contract"
	"github.com/syntopica/worker/internal/link"
	"github.com/syntopica/worker/internal/redact"
)

// attempt is one leased task on its way through its runner.
type attempt struct {
	loop     *Loop
	lease    *contract.Lease
	profile  *config.TaskProfile
	executor contract.Executor
	started  time.Time
}

// run runs the task end to end and returns the report to complete with, or
// nil when the attempt was fenced out.
//
// The scratch directory (workspace, schema, streams) is removed however the
// run ends, so no task input outlives its attempt on disk.
func (a *attempt) run(ctx context.Context) *link.Completion {
	var task taskInput
	if json.Unmarshal(a.lease.Input, &task) != nil {
		return a.refuse(nodeError)
	}

	scratch, err := os.MkdirTemp("", "worker-task-")
	if err != nil {
		return a.refuse(nodeError)
	}

	defer func() { _ = os.RemoveAll(scratch) }()

	call, why := a.prepare(scratch, &task)
	if why != "" {
		return a.refuse(why)
	}

	stdout, stderr, err := streams(scratch)
	if err != nil {
		return a.refuse(nodeError)
	}

	if ctx.Err() != nil {
		_, _ = stdout.Close(), stderr.Close()
		return a.report(preempted, nodeShutdown)
	}
	runner, err := startProcess(a.profile, call, filepath.Join(scratch, "workspace"), stdout, stderr)
	if err != nil {
		_, _ = stdout.Close(), stderr.Close()

		return a.refuse(runnerFailed)
	}

	return a.collect(ctx, runner, stdout, stderr, call.answerFile)
}

func (a *attempt) collect(ctx context.Context, runner *process, stdout, stderr *os.File, answerFile string) *link.Completion {
	how := runner.supervise(ctx, a.profile.Timeout, a.loop.Beat, func() bool {
		return a.loop.Link.Heartbeat(ctx, a.lease.AttemptID, a.lease.Generation, false)
	})
	_, _ = stdout.Close(), stderr.Close()

	switch how {
	case fencedOut:
		return nil
	case overran:
		return a.report(failed, timedOut)
	case stopped:
		return a.report(preempted, nodeShutdown)
	case exited:
	}

	return a.judge(&finished{
		cli:    runnerFor(a.profile.Runner),
		exit:   runner.exit(),
		stdout: read(stdout.Name()),
		stderr: read(stderr.Name()),
	}, answerFile)
}

// prepare builds the workspace and the runner's call. Every runner is a
// remote model, so credentials are redacted from the prompt first; input
// files are passed as they are.
func (a *attempt) prepare(scratch string, task *taskInput) (*invocation, reason) {
	work := &job{
		profile:   a.profile,
		prompt:    redact.Credentials(task.Prompt),
		workspace: filepath.Join(scratch, "workspace"),
		scratch:   scratch,
		schema:    task.schema(),
	}

	if os.Mkdir(work.workspace, privateDir) != nil {
		return nil, nodeError
	}

	if why := buildWorkspace(a.profile.InputRoot, task.Inputs, work.workspace, a.profile.DeniedRoot); why != "" {
		return nil, why
	}

	if work.schema != "" {
		work.schemaFile = filepath.Join(scratch, "output-schema.json")

		if os.WriteFile(work.schemaFile, []byte(work.schema), privateFile) != nil {
			return nil, nodeError
		}
	}

	call := runnerFor(a.profile.Runner).invoke(work)

	return &call, ""
}

// judge reads the runner's answer and reports the run.
func (a *attempt) judge(done *finished, answerFile string) *link.Completion {
	if answerFile == "" {
		done.answer, done.answered = done.cli.answer(done.stdout)
	} else if text := read(answerFile); strings.TrimSpace(text) != "" {
		done.answer, done.answered = text, true
	}

	how, why, output := done.verdict()
	if why == runnerFailed {
		// Executor metadata only: runner text can repeat private task content.
		a.loop.logf("worker: runner_failed %s:%s\n", a.profile.Runner, a.profile.Model)
	}

	report := a.report(how, why)
	report.Output, report.Usage = output, done.usage(why)

	return report
}

// refuse reports an attempt that never ran its runner.
func (a *attempt) refuse(why reason) *link.Completion {
	report := a.report(failed, why)
	report.WallS = 0

	return report
}

// report is the attempt's completion so far.
func (a *attempt) report(how outcome, why reason) *link.Completion {
	report := &link.Completion{
		Outcome:  string(how),
		Usage:    map[string]any{},
		Executor: a.executor,
		WallS:    a.loop.Now().Sub(a.started).Seconds(),
	}

	if why != "" {
		report.ErrorCode = new(string(why))
	}

	return report
}

// streams creates the files the runner's output goes to.
func streams(scratch string) (stdout, stderr *os.File, err error) {
	if stdout, err = os.Create(filepath.Join(scratch, "stdout")); err != nil {
		return nil, nil, fmt.Errorf("create runner stdout: %w", err)
	}

	if stderr, err = os.Create(filepath.Join(scratch, "stderr")); err != nil {
		_ = stdout.Close()

		return nil, nil, fmt.Errorf("create runner stderr: %w", err)
	}

	return stdout, stderr, nil
}

// read is a file's text; a file that cannot be read is empty.
func read(path string) string {
	text, _ := os.ReadFile(path)

	return string(text)
}
