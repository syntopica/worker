package api

import (
	"net/http"
	"strings"
	"time"
)

func discardBody(w http.ResponseWriter, r *http.Request) {
	const (
		maxDiscard = 64 * 1024 * 1024
		chunkSize  = 64 * 1024
		idle       = 500 * time.Millisecond
		total      = 30 * time.Second
	)
	if strings.EqualFold(r.Header.Get("Expect"), "100-continue") {
		return
	}
	controller := http.NewResponseController(w)
	end := time.Now().Add(total)
	buf := make([]byte, chunkSize)
	for left := min(r.ContentLength, maxDiscard); left > 0; {
		now := time.Now()
		if !now.Before(end) {
			return
		}
		deadline := now.Add(idle)
		if end.Before(deadline) {
			deadline = end
		}
		if err := controller.SetReadDeadline(deadline); err != nil {
			return
		}
		n, err := r.Body.Read(buf[:min(left, chunkSize)])
		if err != nil {
			return
		}
		left -= int64(n)
	}
}
