package contract

import "encoding/json"

// Result is one row of GET /v1/results: an output, or a control result that
// carries no output, executor or usage.
type Result struct {
	// Seq orders the results of a queue.
	Seq int64 `json:"seq"`
	// ResultID is what the producer records and acknowledges.
	ResultID string `json:"result_id"`
	// JobID is the job the result belongs to.
	JobID string `json:"job_id"`
	// Control is nil for an output, or names the control result.
	Control *string `json:"control"`
	// Detail depends on Control.
	Detail json.RawMessage `json:"detail"`
	// Output is the job's answer.
	Output json.RawMessage `json:"output"`
	// Executor is what produced the output.
	Executor *Executor `json:"executor"`
	// Usage is what the attempt measured.
	Usage *Usage `json:"usage"`
}
