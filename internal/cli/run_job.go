package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/syntopica/worker/internal/client"
)

type runJob struct {
	State        string     `json:"state"`
	Result       *runResult `json:"result"`
	CoolingUntil *float64   `json:"cooling_until"`
}

func readRunJob(ctx context.Context, cl *client.Client, id string) (runJob, error) {
	raw, err := cl.JobDocument(ctx, id)
	if err != nil {
		return runJob{}, err
	}
	var job runJob
	if err := json.Unmarshal(raw, &job); err != nil {
		return runJob{}, fmt.Errorf("read job state: %w", err)
	}
	return job, nil
}
