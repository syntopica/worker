package jobs

// JobStatus is one producer-owned job and its latest unacknowledged result.
type JobStatus struct {
	ID           string     `json:"id"`
	State        string     `json:"state"`
	Error        *string    `json:"error"`
	Result       *JobResult `json:"result"`
	CoolingUntil *float64   `json:"cooling_until"`
}
