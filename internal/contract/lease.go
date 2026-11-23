package contract

import "encoding/json"

// Lease is the answer to POST /v1/leases when a job is eligible.
type Lease struct {
	// JobID is the leased job.
	JobID string `json:"job_id"`
	// AttemptID fences every heartbeat and completion of this attempt.
	AttemptID string `json:"attempt_id"`
	// Generation is the fencing generation of the lease.
	Generation int64 `json:"generation"`
	// Model is the model the coordinator resolved for the attempt.
	Model string `json:"model"`
	// Input is the job's payload.
	Input json.RawMessage `json:"input"`
	// RunWhen is "idle" or "active_ok".
	RunWhen string `json:"run_when"`
	// TTLSeconds is how long the lease lives without a heartbeat.
	TTLSeconds float64 `json:"ttl_s"`
	// Kind is "inference" or "task".
	Kind string `json:"kind"`
	// Queue is the job's queue.
	Queue string `json:"queue"`
	// Privacy is the job's privacy class.
	Privacy string `json:"privacy"`
	// Pinned says the model was fixed by an override rather than by routing.
	Pinned bool `json:"pinned"`
}
