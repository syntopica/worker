package jobs

import (
	"net/http"

	"github.com/syntopica/worker/internal/fault"
)

// bad refuses a malformed request with 400 and a contract code.
func bad(code string) error {
	return fault.New(http.StatusBadRequest, code)
}
