package contract

// Submitted is the answer to a submission.
type Submitted struct {
	// ID is the job's id.
	ID string `json:"id"`
	// Created is false when the idempotency key matched an existing job.
	Created bool `json:"created"`
}
