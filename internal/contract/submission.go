package contract

import "encoding/json"

// Submission is the body of POST /v1/jobs.
type Submission struct {
	// Contract is the contract version, always Version.
	Contract int `json:"contract"`
	// Kind is "inference" or "task".
	Kind string `json:"kind"`
	// Queue is the producer's queue, "<producer>.<name>".
	Queue string `json:"queue"`
	// IdempotencyKey identifies the job: the same key and body is the same job.
	IdempotencyKey string `json:"idempotency_key"`
	// Priority runs from 0 to 100; higher is leased first.
	Priority int `json:"priority"`
	// Privacy is the privacy class that bounds where the job may run.
	Privacy string `json:"privacy"`
	// Deadline is the epoch second after which the job expires, if any.
	Deadline *float64 `json:"deadline"`
	// MaxAttempts bounds how many leases the job may consume.
	MaxAttempts int `json:"max_attempts"`
	// Tier is "basic" or "strong"; empty means basic.
	Tier string `json:"tier,omitempty"`
	// ParentID names the job a split child replaces.
	ParentID string `json:"parent_id,omitempty"`
	// Requirements is set for inference jobs only.
	Requirements *Requirements `json:"requirements,omitempty"`
	// Input is the job's payload; its shape depends on Kind.
	Input json.RawMessage `json:"input"`
}
