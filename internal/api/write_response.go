package api

import (
	"net/http"
	"strconv"
	"time"
)

func writeResponse(w http.ResponseWriter, status int, data []byte) {
	const responseTimeout = 30 * time.Second
	_ = http.NewResponseController(w).SetWriteDeadline(time.Now().Add(responseTimeout))
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(status)
	_, _ = w.Write(data)
}
