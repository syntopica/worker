package client

import (
	"context"
	"encoding/json"
	"net/http"
)

// JobDocument returns the coordinator's job report without interpreting it.
func (c *Client) JobDocument(ctx context.Context, id string) (json.RawMessage, error) {
	var raw json.RawMessage
	err := c.call(ctx, http.MethodGet, jobPath(id), nil, &raw)
	return raw, err
}
