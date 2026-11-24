package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/policy"
	"github.com/syntopica/worker/internal/store"
)

// onDemandPayloads leaves files and self-evaluations out of an override.
func onDemandPayloads(
	ctx context.Context, tx store.Conn, candidates []policy.Candidate, profile *config.TaskProfile,
) ([]policy.Candidate, error) {
	kept := candidates[:0]
	for index := range candidates {
		candidate := &candidates[index]
		if candidate.Kind == kindTask {
			input, err := loadJobInput(ctx, tx, candidate.JobID)
			if err != nil {
				return nil, err
			}
			if truthy(input[keyInputs]) || candidate.Pin == pinJudge && judgesProfile(input, profile) {
				continue
			}
		}
		kept = append(kept, *candidate)
	}
	return kept, nil
}

func judgesProfile(input map[string]any, profile *config.TaskProfile) bool {
	for _, value := range object(input["labels"]) {
		label := object(value)
		if label["provider"] == profile.Runner && label["model"] == profile.Model {
			return true
		}
	}
	return false
}
