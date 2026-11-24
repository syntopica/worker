package fault_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/syntopica/worker/internal/fault"
)

func TestARefusalKeepsItsStatusAndCodeThroughWrapping(t *testing.T) {
	t.Parallel()

	wrapped := fmt.Errorf("submit: %w", fault.New(http.StatusConflict, "stale_attempt"))

	refusal, ok := errors.AsType[*fault.Error](wrapped)
	if !ok || refusal.Status != http.StatusConflict || refusal.Code != "stale_attempt" {
		t.Fatalf("refusal = %+v", refusal)
	}

	if refusal.Error() != "409 stale_attempt" {
		t.Fatalf("message = %q", refusal.Error())
	}
}
