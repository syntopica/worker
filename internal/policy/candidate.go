package policy

// Candidate is a queued job as the scheduler sees it: no payload, only what
// ranking needs.
type Candidate struct {
	JobID    string
	Queue    string
	Model    string
	Priority int
	Privacy  string
	Created  float64
	Parked   bool
	// ParkedMinIdle is the current idle, in seconds, a parked job needs
	// before it runs.
	ParkedMinIdle float64
	// Kind is "inference" or "task".
	Kind string
	// Pin names the only executor a shadow member may run on; empty means
	// none.
	Pin string
	// Error is the code the job was last handed back with.
	Error string
	// ExcludeProviders are the executor providers that must never run it.
	ExcludeProviders []string
}

// under is the candidate carrying the model or profile it would run on.
func (c *Candidate) under(model string) Candidate {
	routed := *c
	routed.Model = model
	return routed
}
