package config

import "time"

// QueuePolicy is how one queue is scheduled and routed.
type QueuePolicy struct {
	// Name is the queue's name.
	Name string
	// RunWhen is "idle" or "active_ok".
	RunWhen string
	// Weight is the queue's share in fair scheduling.
	Weight int
	// MaxOutstanding bounds the queue's unfinished jobs.
	MaxOutstanding int
	// RetentionDays is how long a finished job is kept.
	RetentionDays int
	// UnackedTTLHours is how long an unfetched result is kept.
	UnackedTTLHours int
	// ParkedMinIdle is the idle time a parked job needs before it runs.
	ParkedMinIdle time.Duration
	// MaxModelAge is how long a queue's model may stay resident unused.
	MaxModelAge time.Duration
	// Profiles holds the task profiles the queue grants.
	Profiles []string
	// Fallbacks maps a granted profile to the profiles, in order, a task may
	// run under while its own runner rests after a quota wall.
	Fallbacks map[string][]string
	// Tiers maps a quality tier to what it runs on.
	Tiers map[string]TierRoute
	// Runner sends the queue's inference jobs to a task profile first; nil
	// means it does not.
	Runner *RunnerRoute
	// LocalAfter holds inference jobs off the local model until they are
	// that old while a remote route could take them.
	LocalAfter time.Duration
	// Shadow samples answers for judged comparison; nil means none.
	Shadow *ShadowPolicy
	// OpenRouter lets the queue's inference jobs escalate to free remote
	// endpoints; nil means they never do.
	OpenRouter *OpenRouterRoute
}
