package api

import (
	"net/http"

	"github.com/syntopica/worker/internal/fault"
)

func badRequest() *fault.Error {
	return fault.New(http.StatusBadRequest, "bad_request")
}
