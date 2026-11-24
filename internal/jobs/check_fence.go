package jobs

import (
	"context"
	"net/http"

	"github.com/syntopica/worker/internal/fault"
	"github.com/syntopica/worker/internal/store"
)

func checkFence(ctx context.Context, tx store.Conn, attempt string, generation int64, node string) (string, error) {
	var id, state string
	found, err := queryRow(ctx, tx, "fenced_job", params{keyAttempt: attempt, keyGeneration: generation, keyNode: node}, &id, &state)
	if err != nil {
		return "", err
	}
	if !found || !liveState(state) {
		return "", fault.New(http.StatusConflict, "stale_attempt")
	}
	return id, nil
}
