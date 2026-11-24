package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/store"
)

func addOutput(ctx context.Context, tx store.Conn, job *jobRow, report *CompletionReport, now float64) error {
	output, err := encodeValue(report.Output)
	if err != nil {
		return errInvalidOutput
	}
	executor, err := encodeValue(report.Executor)
	if err != nil {
		return errInvalidOutput
	}
	usage, err := encodeValue(report.Usage)
	if err != nil {
		return errInvalidOutput
	}
	return run(ctx, tx, "add_output", params{
		keyResultID: newID(), keyID: job.ID, keyProducer: job.Producer, keyQueue: job.Queue,
		keyExecutor: executor, keyUsage: usage, keyNow: now, keyBody: output,
	})
}
