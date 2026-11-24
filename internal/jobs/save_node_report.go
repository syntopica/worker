package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/store"
)

// SaveNodeReport records a node's latest host metadata.
func SaveNodeReport(ctx context.Context, s *store.Store, node, report string, now float64) error {
	return run(ctx, s, "save_node_report", params{keyNode: node, "report": report, keyNow: now})
}
