package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/store"
)

// JobState returns the state of a job an administrator is acting on.
func JobState(ctx context.Context, conn store.Conn, id string) (string, error) {
	var state string
	found, err := queryRow(ctx, conn, "job_state", params{keyID: id}, &state)
	if err != nil {
		return "", err
	}
	if !found {
		return "", notFound()
	}
	return state, nil
}
