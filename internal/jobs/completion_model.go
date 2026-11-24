package jobs

import (
	"regexp"

	"github.com/syntopica/worker/internal/config"
)

const maxRemoteModelLength = 160

var remoteModelID = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+(:[A-Za-z0-9_.-]+)?$`)

func completionModel(cfg *config.Config, job *jobRow, provider, model string) (known, matches bool) {
	switch provider {
	case executorOllama:
		return true, model == job.Model || model == cfg.Models[job.Model].Name
	case kindOpenRouter:
		return true, len(model) <= maxRemoteModelLength && remoteModelID.MatchString(model)
	case executorRunner:
		return true, model == ""
	}
	for name := range cfg.Profiles {
		profile := cfg.Profiles[name]
		if profile.Runner == provider {
			known = true
			if profile.Model == model {
				return true, true
			}
		}
	}
	return known, false
}
