package contract_test

import (
	"testing"

	"github.com/syntopica/worker/internal/contract"
)

func TestRefusalMatchesFixture(t *testing.T) {
	t.Parallel()

	var refusal contract.Refusal

	roundTrip(t, "error_stale_attempt.json", &refusal)

	if refusal.Error != "stale_attempt" {
		t.Errorf("error = %q, want stale_attempt", refusal.Error)
	}
}
