package policy

import (
	"slices"

	"github.com/syntopica/worker/internal/config"
)

// PrivacyAllows reports whether a privacy class may use an executor on a node
// of a trust class. Both must be permitted; absence denies.
func PrivacyAllows(cfg *config.Config, privacy, executor, trust string) bool {
	return slices.Contains(cfg.Privacy[privacy], executor) && slices.Contains(cfg.Trust[privacy], trust)
}
