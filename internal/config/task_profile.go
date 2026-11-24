package config

import "time"

// TaskProfile is an authorized way to run one runner over one input root
// (spec 8).
type TaskProfile struct {
	// Name is the profile's name, which a task lease carries as its model.
	Name string
	// Runner is the runner CLI: "codex", "agy", "cursor", "claude" or
	// "max-lane".
	Runner string
	// Model is the model the runner is asked for; empty means its default.
	Model string
	// Reasoning is the reasoning effort, where the runner takes one.
	Reasoning string
	// Privacy holds the privacy classes the profile may run.
	Privacy []string
	// Timeout is how long one run may take.
	Timeout time.Duration
	// Nodes holds the nodes that may run the profile; nil means any node.
	Nodes []string
	// InputRoot is the directory task inputs are copied from; empty means
	// the profile takes no inputs.
	InputRoot string
	// Command is the runner executable; empty means its name on PATH.
	Command string
	// DeniedRoot is never copied into a workspace, even inside InputRoot:
	// the worker's own state, which holds the bearer tokens.
	DeniedRoot string
	// EnvUnset holds the variables removed from the runner's environment,
	// such as an API key that would bill a call the subscription login
	// should carry.
	EnvUnset []string
	// OnDemand keeps the profile out of every route, fallback and pin: it
	// runs only when a task loop asks for it by name.
	OnDemand bool
	// Env holds the variables set after EnvUnset, in order, such as the
	// configuration directory that selects a runner's account.
	Env [][2]string
}
