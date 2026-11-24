package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, value any) error {
	w.Header().Set("Content-Type", "application/json")
	if status == http.StatusNoContent {
		w.WriteHeader(status)
		return nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode API response: %w", err)
	}
	writeResponse(w, status, data)
	return nil
}
