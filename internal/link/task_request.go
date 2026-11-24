package link

// TaskRequest is what a task loop tells the coordinator when it asks for
// work. The last three fields ask on demand and are left out otherwise.
type TaskRequest struct {
	// Node is the asking node; LeaseTask sets it.
	Node string `json:"node"`
	// Kind is "task"; LeaseTask sets it.
	Kind string `json:"kind"`
	// UserActive says whether the user is below the idle threshold.
	UserActive bool `json:"user_active"`
	// CurrentIdleS is the idle time in seconds.
	CurrentIdleS float64 `json:"current_idle_s"`
	// Profile asks for jobs of one on-demand profile.
	Profile string `json:"profile,omitempty"`
	// OverridePins also takes jobs pinned to another profile.
	OverridePins bool `json:"override_pins,omitempty"`
	// CreatedBefore takes only jobs created before this Unix time.
	CreatedBefore *float64 `json:"created_before,omitempty"`
}
