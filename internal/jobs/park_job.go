package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/store"
)

const parkP90Multiple = 1.5

func parkJob(ctx context.Context, tx store.Conn, id, queue string, floor, now float64) error {
	var p90 float64
	if _, err := queryRow(ctx, tx, "queue_p90_runtime", params{keyQueue: queue}, &p90); err != nil {
		return err
	}
	return run(ctx, tx, "park_job", params{keyID: id, keyNow: now, "idle": max(floor, parkP90Multiple*p90)})
}
