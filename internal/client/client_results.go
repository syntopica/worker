package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/syntopica/worker/internal/contract"
)

// Results returns the unacknowledged results a query selects, in Seq order.
func (c *Client) Results(ctx context.Context, query ResultsQuery) ([]contract.Result, error) {
	var page struct {
		Results []contract.Result `json:"results"`
	}
	err := c.resultPage(ctx, query, &page)
	return page.Results, err
}

// ResultDocuments returns rows exactly as the coordinator sent them, for
// consumers that pass results through without interpreting their fields.
func (c *Client) ResultDocuments(ctx context.Context, query ResultsQuery) ([]json.RawMessage, error) {
	var page struct {
		Results []json.RawMessage `json:"results"`
	}
	err := c.resultPage(ctx, query, &page)
	return page.Results, err
}

func (c *Client) resultPage(ctx context.Context, query ResultsQuery, out any) error {
	const (
		defaultLimit = 50
		pollMargin   = 10 * time.Second // Outlive the server's own long poll.
	)
	limit := query.Limit
	if limit == 0 {
		limit = defaultLimit
	}

	values := url.Values{
		"queue": {query.Queue},
		"after": {strconv.FormatInt(query.After, 10)},
		"limit": {strconv.Itoa(limit)},
		"wait":  {strconv.FormatFloat(query.Wait.Seconds(), 'f', -1, 64)},
	}

	ctx, cancel := context.WithTimeout(ctx, max(0, query.Wait)+pollMargin)
	defer cancel()

	return c.call(ctx, http.MethodGet, "/v1/results?"+values.Encode(), nil, out)
}
