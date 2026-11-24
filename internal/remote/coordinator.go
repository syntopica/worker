package remote

import (
	"context"

	"github.com/syntopica/worker/internal/contract"
	"github.com/syntopica/worker/internal/link"
)

// Coordinator is what the loop needs from its link to the coordinator.
type Coordinator interface {
	// LeaseRemote asks for an escalated inference job; nil means none.
	LeaseRemote(ctx context.Context) *contract.Lease
	// Heartbeat keeps an attempt's lease; false means it was fenced out.
	Heartbeat(ctx context.Context, attemptID string, generation int64, draining bool) bool
	// Complete reports how an attempt ended.
	Complete(ctx context.Context, attemptID string, generation int64, report *link.Completion)
}
