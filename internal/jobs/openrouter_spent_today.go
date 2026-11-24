package jobs

import (
	"context"
	"math"

	"github.com/syntopica/worker/internal/store"
)

func openrouterSpentToday(ctx context.Context, tx store.Conn, now float64) (map[string]int, error) {
	return queryCounts(ctx, tx, "openrouter_spent_today", params{"midnight": math.Floor(now/daySeconds) * daySeconds})
}
