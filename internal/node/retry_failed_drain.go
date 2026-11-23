package node

import (
	"context"
	"slices"
	"time"

	"github.com/syntopica/worker/internal/ollama"
)

// drainRetry is how often a failed drain's unload is sent again.
const drainRetry = 5 * time.Minute

// retryFailedDrain retries the unload of a model whose drain failed, on a
// slow timer.
//
// A drain whose unload and restart both failed leaves the model listed and
// the node blocked; nothing else would try again. The block itself clears in
// step when /api/ps stops listing the model.
func retryFailedDrain(ctx context.Context, backend ollama.Backend, m *memory, resident []string, now time.Time) {
	if m.failedModel == "" || !slices.Contains(resident, m.failedModel) || now.Before(m.drainRetryAt) {
		return
	}

	m.drainRetryAt = now.Add(drainRetry)
	backend.Unload(ctx, m.failedModel)
}
