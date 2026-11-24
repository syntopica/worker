package client

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// AdminReport returns the JSON body from a coordinator report endpoint.
func (c *Client) AdminReport(ctx context.Context, path string) (json.RawMessage, error) {
	const timeout = 10 * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var raw json.RawMessage
	err := c.call(ctx, http.MethodGet, path, nil, &raw)
	return raw, err
}
