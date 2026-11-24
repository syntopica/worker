package policy

import (
	"slices"

	"github.com/syntopica/worker/internal/config"
)

// onNode reports whether a profile may run on a node; no list means any.
func onNode(profile *config.TaskProfile, node string) bool {
	return profile.Nodes == nil || slices.Contains(profile.Nodes, node)
}
