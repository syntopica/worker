package api

import (
	"net/http"
	"strconv"
)

func writeRefusal(w http.ResponseWriter, status int, code string) {
	writeResponse(w, status, []byte(`{"error": `+strconv.Quote(code)+`}`))
}
