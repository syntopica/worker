package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/store"
)

// ReadAudit returns content-free operator actions since the given time, newest first.
func ReadAudit(ctx context.Context, s *store.Store, since float64) ([]map[string]any, error) {
	return queryMaps(ctx, s, "read_audit", params{keySince: since})
}
