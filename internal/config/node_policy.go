package config

import "time"

// NodePolicy is what one machine may run and how it reaches its model
// server.
type NodePolicy struct {
	// Name is the node's name.
	Name string
	// Trust is "owner", "server" or "guest".
	Trust string
	// IdleThreshold is the idle time below which the user counts as active.
	IdleThreshold time.Duration
	// MemoryBudgetGB is the memory the local model may use.
	MemoryBudgetGB float64
	// OllamaURL is the base URL of the node's model server.
	OllamaURL string
	// OllamaLaunchdLabel is the LaunchAgent restarted as the last drain step.
	OllamaLaunchdLabel string
	// MinFreePct is the free memory below which a kernel pressure warning
	// counts as pressure on this machine.
	MinFreePct float64
	// MaxLoad is the load average above which a Linux server counts as in
	// use; nil means no limit.
	MaxLoad *float64
	// LocalIdleOnly keeps the local model off an owner's workstation while
	// the owner is active, active_ok queues included: no local lease below
	// the idle threshold, every local attempt released on input, and the
	// node's own model unloaded rather than left resident for its KeepAlive.
	LocalIdleOnly bool
	// OpenRouterKeyFile is the file, local to the node, holding its
	// OpenRouter key; empty means no remote loop.
	OpenRouterKeyFile string
	// RemoteSlots is how many OpenRouter calls the remote loop keeps in
	// flight.
	RemoteSlots int
	// TaskSlots is how many runner tasks the task loop runs at once.
	TaskSlots int
	// CoordinatorURL is how the node reaches the coordinator when that is
	// not the configured listen address.
	CoordinatorURL string
}
