package client

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/syntopica/worker/internal/contract"
)

// SubmitDocument admits a job from its original JSON, leaving server-side
// defaults and optional or future contract fields intact.
func (c *Client) SubmitDocument(ctx context.Context, job json.RawMessage) (contract.Submitted, error) {
	var submitted contract.Submitted
	err := c.call(ctx, http.MethodPost, "/v1/jobs", job, &submitted)
	return submitted, err
}
