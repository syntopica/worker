package config

import "time"

// Config is the instance configuration, parsed and defaulted.
type Config struct {
	// ListenHost is the loopback host the coordinator binds.
	ListenHost string
	// ListenPort is the port the coordinator binds.
	ListenPort int
	// Producers maps a producer to the queues it is granted.
	Producers map[string][]string
	// MaxPayloadBytes bounds one submitted job.
	MaxPayloadBytes int
	// MaxSplit bounds the parts one job may be split into.
	MaxSplit int
	// MaxParkedRuns bounds how often a parked job is tried again.
	MaxParkedRuns int
	// SplitAfterPreemptions is the preemption count that splits a job.
	SplitAfterPreemptions int
	// RunnerCooldown is how long a runner rests after a quota wall.
	RunnerCooldown map[string]time.Duration
	// RunnerMaxConcurrent caps a runner's live attempts across all nodes.
	RunnerMaxConcurrent map[string]int
	// Models maps a managed model's name to its pin.
	Models map[string]ModelPin
	// Nodes maps a node's name to its policy.
	Nodes map[string]NodePolicy
	// Queues maps a queue's name to its policy.
	Queues map[string]QueuePolicy
	// Profiles maps a task profile's name to the profile.
	Profiles map[string]TaskProfile
	// RunnerQuota maps a runner to where its quota is read from.
	RunnerQuota map[string]RunnerQuota
	// RunnerModelWindows maps a runner to its rate window labels, and each
	// label to the model-name fragments that window meters.
	RunnerModelWindows map[string]map[string][]string
	// Privacy maps a privacy class to the executors it may use.
	Privacy map[string][]string
	// Trust maps a privacy class to the node trust classes it may run on.
	Trust map[string][]string
}
