package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/store"
)

func submitJudgesBatch(ctx context.Context, tx store.Conn, cfg *config.Config, now float64, limit int) (int, error) {
	groups, err := queryStrings(ctx, tx, "groups_ready_for_judging", params{keyLimit: limit})
	if err != nil {
		return 0, err
	}
	for _, group := range groups {
		if err := submitJudge(ctx, tx, cfg, group, now); err != nil {
			return 0, err
		}
	}
	return len(groups), nil
}
