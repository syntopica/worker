package config

// Config is the instance configuration, parsed and defaulted.
type Config struct {
	// Models maps a managed model's name to its pin.
	Models map[string]ModelPin
	// Nodes maps a node's name to its policy.
	Nodes map[string]NodePolicy
	// Privacy maps a privacy class to the executors it may use.
	Privacy map[string][]string
	// Trust maps a privacy class to the node trust classes it may run on.
	Trust map[string][]string
}
