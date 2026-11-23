package contract_test

import (
	"testing"

	"github.com/syntopica/worker/internal/contract"
)

func TestSubmissionMatchesFixtures(t *testing.T) {
	t.Parallel()

	for name, kind := range map[string]string{
		"submit_inference.json": "inference",
		"submit_task.json":      "task",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var submission contract.Submission

			roundTrip(t, name, &submission)

			if submission.Contract != contract.Version || submission.Kind != kind {
				t.Errorf("contract/kind = %d/%q, want %d/%q", submission.Contract, submission.Kind, contract.Version, kind)
			}

			if (submission.Requirements != nil) != (kind == "inference") {
				t.Errorf("requirements present = %t for a %s job", submission.Requirements != nil, kind)
			}
		})
	}
}
