package jobs

// submission is a validated submit body (contract v1).
type submission struct {
	Kind        string
	Queue       string
	Key         string
	Privacy     string
	Tier        string
	Priority    int
	MaxAttempts int
	Models      []string
	Input       map[string]any
	ParentID    *string
	Deadline    *float64
	Exclude     []string
}
