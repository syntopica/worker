package policy

// LeaseRequest is what a node reports when it asks for work.
type LeaseRequest struct {
	Node string
	// ResidentModel is the model loaded on the node; empty means none.
	ResidentModel string
	UserActive    bool
	// FreeGB is the node budget minus the resident model's footprint.
	FreeGB float64
	// CurrentIdle is the node's idle time in seconds.
	CurrentIdle float64
	// Kind is the job kind the asking loop runs.
	Kind string
	// Profile names the on-demand profile a task loop was started for.
	Profile string
	// CreatedBefore keeps the lease to jobs older than it; nil means all.
	CreatedBefore *float64
	OverridePins  bool
}
