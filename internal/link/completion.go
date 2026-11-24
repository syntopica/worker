package link

import "github.com/syntopica/worker/internal/contract"

// Completion is what a node reports when an attempt ends. ErrorCode is
// allowlisted: never a provider body or generated text.
type Completion struct {
	// Outcome is "succeeded", "failed" or "preempted".
	Outcome string `json:"outcome"`
	// Output is the answer of a succeeded attempt, else nil.
	Output map[string]any `json:"output"`
	// Usage is what the attempt measured; empty when it measured nothing.
	Usage map[string]any `json:"usage"`
	// Executor is what ran the attempt.
	Executor contract.Executor `json:"executor"`
	// ErrorCode says why an attempt did not succeed.
	ErrorCode *string `json:"error_code"`
	// WallS is the attempt's wall time in seconds.
	WallS float64 `json:"wall_s"`
	// Generation fences the completion; Complete sets it.
	Generation int64 `json:"generation"`
}
