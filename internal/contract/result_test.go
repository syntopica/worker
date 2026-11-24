package contract_test

import (
	"testing"

	"github.com/syntopica/worker/internal/contract"
)

func TestResultMatchesFixtures(t *testing.T) {
	t.Parallel()

	for name, control := range map[string]string{
		"result_succeeded.json":       "",
		"result_openrouter.json":      "",
		"result_failed.json":          "failed",
		"result_split_requested.json": "split_requested",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var result contract.Result

			roundTrip(t, name, &result)

			got := ""
			if result.Control != nil {
				got = *result.Control
			}

			if got != control {
				t.Errorf("control = %q, want %q", got, control)
			}

			if (result.Executor == nil) != (control != "") {
				t.Errorf("executor present = %t for control %q", result.Executor != nil, control)
			}

			checkResultCost(t, name, result.Usage)
		})
	}
}

func checkResultCost(t *testing.T, fixture string, usage *contract.Usage) {
	t.Helper()

	if fixture == "result_openrouter.json" {
		if usage == nil || usage.CostUSD == nil || *usage.CostUSD != 0 {
			t.Error("usage did not preserve the reported zero cost")
		}
		return
	}

	if usage != nil && usage.CostUSD != nil {
		t.Errorf("usage unexpectedly has cost_usd = %v", *usage.CostUSD)
	}
}
