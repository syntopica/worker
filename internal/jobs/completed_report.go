package jobs

import "github.com/syntopica/worker/internal/config"

// completedReport keeps the fixed vocabulary and metadata fields. Free-form
// output may only be written to the payload database.
func completedReport(cfg *config.Config, job *jobRow, report *CompletionReport) CompletionReport {
	settlement := *report
	settlement.Executor = completionExecutor(cfg, job, report.Executor)
	settlement.Usage = map[string]any{}
	for _, key := range []string{keyTokensIn, keyTokensOut, keyCostUSD, keyWallReset} {
		if value := sqlValue(report.Usage[key]); value != nil {
			settlement.Usage[key] = value
		}
	}
	settlement.ErrorCode = knownErrorCode(report.ErrorCode)
	switch settlement.Outcome {
	case stateSucceeded, stateFailed, outcomePreempted:
	default:
		settlement.Outcome, settlement.ErrorCode = stateFailed, codeExecutorError
	}
	return settlement
}
