package api

import (
	"errors"
	"io"
	"net/http"

	"github.com/syntopica/worker/internal/fault"
)

func readBody(w http.ResponseWriter, r *http.Request, limit int) ([]byte, error) {
	if r.ContentLength > int64(limit) {
		w.Header().Set("Connection", "close")
		discardBody(w, r)
		return nil, fault.New(http.StatusRequestEntityTooLarge, "payload_too_large")
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, int64(limit)))
	var size *http.MaxBytesError
	switch {
	case errors.As(err, &size):
		return nil, fault.New(http.StatusRequestEntityTooLarge, "payload_too_large")
	case err != nil:
		return nil, badRequest()
	default:
		return body, nil
	}
}
