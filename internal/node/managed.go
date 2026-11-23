package node

import "github.com/syntopica/worker/internal/config"

// managed reports whether the configuration pins a model.
func managed(cfg *config.Config, model string) bool {
	_, pinned := cfg.Models[model]

	return pinned
}
