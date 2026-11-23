package contract

// Job is the answer to GET /v1/jobs/{id}.
type Job struct {
	// ID is the job's id.
	ID string `json:"id"`
	// State is the job's lifecycle state.
	State string `json:"state"`
	// Error is the code the job failed with, if it did.
	Error *string `json:"error"`
	// Result is the latest unacknowledged result, if there is one.
	Result *Result `json:"result"`
	// CoolingUntil is the epoch second a resting runner's cooldown ends.
	CoolingUntil *float64 `json:"cooling_until"`
}
