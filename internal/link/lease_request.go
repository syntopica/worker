package link

// LeaseRequest is what an inference node tells the coordinator when it asks
// for work.
type LeaseRequest struct {
	// Node is the asking node; Lease sets it.
	Node string `json:"node"`
	// ResidentModel is the managed model in memory, or nil.
	ResidentModel *string `json:"resident_model"`
	// UserActive says whether the user is below the idle threshold.
	UserActive bool `json:"user_active"`
	// FreeGB is the node budget minus the resident model's footprint.
	FreeGB float64 `json:"free_gb"`
	// CurrentIdleS is the idle time in seconds.
	CurrentIdleS float64 `json:"current_idle_s"`
}
