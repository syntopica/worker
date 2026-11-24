package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/store"
)

func statusQueues(ctx context.Context, s *store.Store, now float64) (map[string]*queueStatus, error) {
	states, err := queryMaps(ctx, s, "status_queues", nil)
	if err != nil {
		return nil, err
	}
	recent, err := queryMaps(ctx, s, "status_recent_attempts", params{keySince: now - hourSeconds})
	if err != nil {
		return nil, err
	}
	queues := map[string]*queueStatus{}
	for _, row := range states {
		status := queueStatusForRow(queues, row)
		state := text(row[keyState])
		status.States[state] = row["n"]
		switch state {
		case stateQueued:
			status.SamplingQueue = row["sampling"]
			if oldest, hasTime := row["oldest"].(float64); hasTime {
				age := now - oldest
				status.OldestQueued = &age
			}
		case stateFailed:
			status.SamplingFail = row["sampling"]
		}
	}
	for _, row := range recent {
		status := queueStatusForRow(queues, row)
		switch row["outcome"] {
		case stateSucceeded:
			status.Done = row["n"]
		case outcomePreempted:
			status.Wasted = row["wall_s"]
		}
	}
	return queues, nil
}

func queueStatusForRow(queues map[string]*queueStatus, row map[string]any) *queueStatus {
	name := text(row[keyQueue])
	if status := queues[name]; status != nil {
		return status
	}
	status := &queueStatus{States: map[string]any{}, Done: 0, Wasted: 0.0, SamplingFail: 0, SamplingQueue: 0}
	queues[name] = status
	return status
}
