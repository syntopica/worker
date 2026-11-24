package config

import "time"

// RunnerRoute is how a queue sends its inference jobs to a runner first.
type RunnerRoute struct {
	// Profile is the task profile that runs the job.
	Profile string
	// After is how long a job waits since it was created.
	After time.Duration
	// Fallbacks are tried in order when Profile cannot run the job now.
	Fallbacks []string
}
