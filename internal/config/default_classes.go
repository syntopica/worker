package config

import "slices"

// defaultClasses is, for each privacy class, the executors it may use and
// the node trust classes it may run on unless the instance overrides it
// (spec 8).
func defaultClasses() (privacy, trust map[string][]string) {
	local := []string{kindOllama, "local-cpu"}
	runners := append(slices.Clone(local), kindRunner)
	owner := []string{"owner"}
	servers := append(slices.Clone(owner), "server")

	privacy, trust = map[string][]string{}, map[string][]string{}

	for _, class := range []struct {
		name             string
		executors, trust []string
	}{
		{"public", append(slices.Clone(local), kindOpenRouter, kindRunner), append(slices.Clone(servers), "guest")},
		{"internal", runners, servers},
		{"personal", local, owner},
		{"mail", local, owner},
		{"secret", local, owner},
	} {
		privacy[class.name], trust[class.name] = class.executors, class.trust
	}

	return privacy, trust
}
