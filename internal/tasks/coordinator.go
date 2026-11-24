package tasks

import (
	"context"

	"github.com/syntopica/worker/internal/contract"
	"github.com/syntopica/worker/internal/link"
)

// Coordinator is what the loop needs from its link to the coordinator.
type Coordinator interface {
	// LeaseTask asks for a task; nil means none.
	LeaseTask(ctx context.Context, request *link.TaskRequest) *contract.Lease
	// Heartbeat keeps an attempt's lease; false means it was fenced out.
	Heartbeat(ctx context.Context, attemptID string, generation int64, draining bool) bool
	// Complete reports how an attempt ended.
	Complete(ctx context.Context, attemptID string, generation int64, report *link.Completion)
	// ReportWalls tells until when each runner's quota is spent.
	ReportWalls(ctx context.Context, walls map[string]float64)
	// Flush retries the completions the link still holds.
	Flush(ctx context.Context)
}
