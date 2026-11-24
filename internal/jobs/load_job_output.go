package jobs

import (
	"context"
	"encoding/json"

	"github.com/syntopica/worker/internal/store"
)

// loadJobOutput returns the latest stored output, if present, from the payload database.
func loadJobOutput(ctx context.Context, s store.Conn, id string) (json.RawMessage, error) {
	var body []byte
	found, err := queryRow(ctx, s, "load_job_output", params{keyID: id}, &body)
	if err != nil || !found {
		return body, err
	}
	if !json.Valid(body) {
		return nil, errInvalidOutput
	}
	return body, nil
}
