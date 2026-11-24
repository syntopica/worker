package cli

import (
	"context"
	"encoding/json"

	"github.com/syntopica/worker/internal/client"
	"github.com/syntopica/worker/internal/contract"
)

func (c *command) producer(ctx context.Context) (int, error) {
	args, err := c.producerArgs()
	if err != nil {
		return usageExit, err
	}
	cl, err := c.client(args.principal)
	if err != nil {
		return usageExit, err
	}
	switch c.name {
	case jobsCommand:
		return 0, c.listResults(ctx, cl, args.queue)
	case cancelCommand:
		state, err := cl.Cancel(ctx, args.values[0])
		if err != nil {
			return failedExit, err
		}
		return 0, writeJSON(c.out, map[string]string{"state": state}, false)
	case rateCommand:
		return 0, c.rate(ctx, cl, args.values)
	default:
		return c.submit(ctx, cl, args)
	}
}

func (c *command) listResults(ctx context.Context, cl *client.Client, queue string) error {
	const limit = 100
	rows, err := cl.ResultDocuments(ctx, client.ResultsQuery{Queue: queue, Limit: limit})
	if err != nil {
		return err
	}
	if rows == nil {
		rows = []json.RawMessage{}
	}
	return writeJSON(c.out, rows, true)
}

func (c *command) rate(ctx context.Context, cl *client.Client, values []string) error {
	if err := cl.Ack(ctx, values[0], contract.Ack{ResultID: values[1], Rating: values[2]}); err != nil {
		return err
	}
	return writeJSON(c.out, map[string]string{"rated": values[2]}, false)
}

func (c *command) submit(ctx context.Context, cl *client.Client, args producerArgs) (int, error) {
	raw, err := c.readJob(args.values[0], c.name == runCommand)
	if err != nil {
		return usageExit, err
	}
	submitted, err := cl.SubmitDocument(ctx, raw)
	if err != nil {
		return failedExit, err
	}
	if c.name == "submit" {
		return 0, writeJSON(c.out, submitted, false)
	}
	return c.waitForResult(ctx, cl, submitted.ID, args)
}
