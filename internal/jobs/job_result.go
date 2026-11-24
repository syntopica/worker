package jobs

import "encoding/json"

// JobResult is the producer's output or control result, in contract v1 form.
type JobResult struct {
	Seq      int64           `json:"seq"`
	ResultID string          `json:"result_id"`
	JobID    string          `json:"job_id"`
	Control  *string         `json:"control"`
	Detail   json.RawMessage `json:"detail"`
	Output   json.RawMessage `json:"output"`
	Executor json.RawMessage `json:"executor"`
	Usage    json.RawMessage `json:"usage"`
}

func (r *JobResult) dest() []any {
	return []any{&r.Seq, &r.ResultID, &r.JobID, &r.Control, (*[]byte)(&r.Detail), (*[]byte)(&r.Output), (*[]byte)(&r.Executor), (*[]byte)(&r.Usage)}
}
