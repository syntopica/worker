package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/url"

	"github.com/syntopica/worker/internal/client"
)

type command struct {
	synopsis    string
	name        string
	args        []string
	in          io.Reader
	out, errOut io.Writer
}

func (c *command) dispatch(ctx context.Context) (int, error) {
	switch c.name {
	case "serve":
		return c.serve(ctx)
	case nodeCommand, tasksCommand, remoteCommand:
		return c.executor(ctx)
	case "status", "nodes", "costs", "quality", "audit":
		return c.report(ctx)
	case "submit", jobsCommand, cancelCommand, rateCommand, runCommand:
		return c.producer(ctx)
	case "token":
		return c.token()
	case "backup":
		return c.backup(ctx)
	}
	_, _ = fmt.Fprintln(c.errOut, usage)
	return usageExit, nil
}

func (c *command) outcome(code int, err error) int {
	switch {
	case err == nil:
		return code
	case errors.Is(err, flag.ErrHelp):
		return 0
	case errors.Is(err, context.Canceled):
		return c.interrupted()
	case errors.Is(err, errUsage):
		return usageExit
	}

	var refusal *client.Error
	var known userError
	var transport *url.Error
	switch {
	case errors.As(err, &refusal):
		safe := refusal.Code
		if !refusalCode.MatchString(safe) {
			safe = unknownValue
		}
		_, _ = fmt.Fprintf(c.errOut, "worker: %d %s\n", refusal.Status, safe)
		return failedExit
	case errors.As(err, &transport):
		_, _ = fmt.Fprintln(c.errOut, "worker: coordinator unreachable")
		return failedExit
	case errors.As(err, &known):
		_, _ = fmt.Fprintln(c.errOut, "worker:", known)
	default:
		_, _ = fmt.Fprintln(c.errOut, "worker:", c.name, "failed")
	}

	return max(code, failedExit)
}

func (c *command) interrupted() int {
	const interruptedExit = 130
	if DrainsOnSignal(c.name) {
		return 0
	}
	return interruptedExit
}
