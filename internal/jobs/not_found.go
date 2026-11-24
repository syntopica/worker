package jobs

import (
	"net/http"

	"github.com/syntopica/worker/internal/fault"
)

func notFound() *fault.Error {
	return fault.New(http.StatusNotFound, codeNotFound)
}
