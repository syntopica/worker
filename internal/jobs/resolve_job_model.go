package jobs

import (
	"net/http"
	"slices"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/fault"
)

// resolveJobModel is the model, or the task profile, a submitted job runs on
// (spec 6, quality tiers). The tier route's choice comes first; a task's
// profile grant and privacy ceiling are checked here too.
func resolveJobModel(cfg *config.Config, sub *submission) (string, error) {
	queue := cfg.Queues[sub.Queue]
	route := queue.Tiers[sub.Tier]
	if sub.Kind == kindTask {
		granted, err := checkTaskGrant(cfg, &queue, sub)
		if mapped, ok := route.Profiles[granted]; ok {
			granted = mapped
		}

		return granted, err
	}
	for _, model := range slices.Concat(route.Models, sub.Models) {
		if _, known := cfg.Models[model]; known {
			return model, nil
		}
	}

	return "", bad("unknown_model")
}

// checkTaskGrant says whether a queue's grant covers a task's profile,
// runner, privacy and inputs (spec 8). A producer may leave the runner to the
// profile, and never names one it does not use, a privacy class the profile
// does not accept, or inputs where the profile has no root.
func checkTaskGrant(cfg *config.Config, queue *config.QueuePolicy, sub *submission) (string, error) {
	name := text(sub.Input[keyProfile])
	profile, known := cfg.Profiles[name]
	runner, named := sub.Input[keyRunner]
	switch {
	case !known || !slices.Contains(queue.Profiles, name):
		return "", fault.New(http.StatusForbidden, "profile_not_granted")
	case named && runner != profile.Runner:
		return "", bad("runner_mismatch")
	case !slices.Contains(profile.Privacy, sub.Privacy):
		return "", fault.New(http.StatusForbidden, "privacy_not_allowed")
	case truthy(sub.Input[keyInputs]) && profile.InputRoot == "":
		return "", bad("inputs_not_allowed")
	}

	return name, nil
}
