package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/syntopica/worker/internal/client"
	"github.com/syntopica/worker/internal/contract"
)

func (c *command) waitForResult(ctx context.Context, cl *client.Client, id string, args producerArgs) (int, error) {
	const (
		parkedExit  = 3
		pendingExit = 4
	)
	deadline := time.Now().Add(args.wait)
	for ctx.Err() == nil {
		job, err := readRunJob(ctx, cl, id)
		if err != nil {
			return failedExit, err
		}
		if code, finished, err := c.collect(ctx, cl, id, job.Result); finished {
			return code, err
		}
		if job.Result == nil && terminal(job.State) {
			_, _ = fmt.Fprintf(c.errOut, "worker job %s is %s with nothing to collect\n", id, job.State)
			return failedExit, nil
		}
		if job.CoolingUntil != nil {
			_, _ = fmt.Fprintf(c.errOut, "worker job %s is parked behind a quota wall\n", id)
			return parkedExit, nil
		}
		left := time.Until(deadline)
		if left <= 0 {
			_, _ = fmt.Fprintf(c.errOut, "worker job %s still %s after %gs\n", id, job.State, args.wait.Seconds())
			return pendingExit, nil
		}
		pause(ctx, min(left, args.poll))
	}
	return failedExit, fmt.Errorf("wait for job: %w", ctx.Err())
}

func (c *command) collect(ctx context.Context, cl *client.Client, id string, result *runResult) (code int, finished bool, err error) {
	if result == nil {
		return 0, false, nil
	}
	split := result.Control != nil && *result.Control == "split_requested"
	if ackErr := cl.Ack(ctx, id, contract.Ack{ResultID: result.ResultID, Decline: split}); ackErr != nil {
		return failedExit, true, ackErr
	}
	if split {
		return 0, false, nil
	}
	code, err = c.printResult(id, result)
	return code, true, err
}
