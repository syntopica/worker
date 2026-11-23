package contract

// Ack is the body of POST /v1/jobs/{id}/ack.
type Ack struct {
	// ResultID is the result being acknowledged.
	ResultID string `json:"result_id"`
	// Decline refuses a split_requested control result, which parks the job.
	Decline bool `json:"decline"`
	// Rating is "good", "edited" or "discarded"; empty leaves it unrated.
	Rating string `json:"rating,omitempty"`
}
