package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/store"
)

func endAttempt(ctx context.Context, tx store.Conn, attempt string, report *CompletionReport, now float64) error {
	return run(ctx, tx, "end_attempt", params{
		keyAttempt: attempt, keyNow: now, "outcome": report.Outcome, keyCode: nullIfEmpty(report.ErrorCode),
		"wall_s": report.WallS, keyTokensIn: report.Usage[keyTokensIn], keyTokensOut: report.Usage[keyTokensOut],
		keyProvider: report.Executor[keyProvider], keyCostUSD: report.Usage[keyCostUSD],
		keyModel: nullIfEmpty(text(report.Executor[keyModel])),
	})
}
