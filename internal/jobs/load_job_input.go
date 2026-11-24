package jobs

import (
	"context"

	"github.com/syntopica/worker/internal/store"
)

// loadJobInput returns nil for an input lost from the payload database.
func loadJobInput(ctx context.Context, tx store.Conn, id string) (map[string]any, error) {
	var body string
	found, err := queryRow(ctx, tx, "load_input", params{keyID: id}, &body)
	if err != nil || !found {
		return nil, err
	}
	value, err := decodeValue([]byte(body))
	input := object(object(value)[keyInput])
	if err != nil || input == nil {
		return nil, errInvalidInput
	}
	return input, nil
}
