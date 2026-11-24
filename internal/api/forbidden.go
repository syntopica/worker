package api

import (
	"net/http"

	"github.com/syntopica/worker/internal/fault"
)

func forbidden() *fault.Error {
	return fault.New(http.StatusForbidden, "forbidden")
}
