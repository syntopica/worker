package jobs

// CompletionReport is the outcome and measurements reported by a node.
type CompletionReport struct {
	Outcome   string
	Output    any
	Usage     map[string]any
	Executor  map[string]any
	ErrorCode string
	WallS     float64
}
