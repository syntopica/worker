package node

import (
	"context"

	"github.com/syntopica/worker/internal/contract"
	"github.com/syntopica/worker/internal/link"
)

// Coordinator is what the node needs from its link to the coordinator.
type Coordinator interface {
	// Lease asks for work; nil means none.
	Lease(ctx context.Context, request link.LeaseRequest) *contract.Lease
	// Heartbeat keeps an attempt's lease; false means it was fenced out.
	Heartbeat(ctx context.Context, attemptID string, generation int64, draining bool) bool
	// Complete reports how an attempt ended.
	Complete(ctx context.Context, attemptID string, generation int64, report *link.Completion)
	// Report sends the host state.
	Report(ctx context.Context, report *link.HostReport)
}
