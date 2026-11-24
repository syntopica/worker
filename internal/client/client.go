package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/syntopica/worker/internal/contract"
)

// Client talks to a coordinator as one producer. It is safe for concurrent
// use.
type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

// New returns a client for the coordinator at baseURL, authenticated by the
// producer's token. A nil httpClient means http.DefaultClient; deadlines come
// from the context of each call.
func New(baseURL, token string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{baseURL: strings.TrimRight(baseURL, "/"), token: token, http: httpClient}
}

// Submit admits a job. Resubmitting the same body returns the same job.
func (c *Client) Submit(ctx context.Context, job *contract.Submission) (contract.Submitted, error) {
	var submitted contract.Submitted

	err := c.call(ctx, http.MethodPost, "/v1/jobs", job, &submitted)

	return submitted, err
}

// Get returns a job's state and its latest unacknowledged result.
func (c *Client) Get(ctx context.Context, jobID string) (contract.Job, error) {
	var job contract.Job

	err := c.call(ctx, http.MethodGet, jobPath(jobID), nil, &job)

	return job, err
}

// Ack acknowledges a result, or declines a split_requested control result.
// Acknowledging the same result twice is harmless.
func (c *Client) Ack(ctx context.Context, jobID string, ack contract.Ack) error {
	return c.call(ctx, http.MethodPost, jobPath(jobID)+"/ack", ack, nil)
}

// Cancel cancels a job and returns the state it ended in.
func (c *Client) Cancel(ctx context.Context, jobID string) (string, error) {
	var job contract.Job

	err := c.call(ctx, http.MethodPost, jobPath(jobID)+"/cancel", struct{}{}, &job)

	return job.State, err
}

// WaitFor polls a job until it has an unacknowledged result, control results
// included, and returns it. It returns the context's error when ctx ends
// first, which is also what happens for a job that will never have one, such
// as a cancelled or already acknowledged job: give ctx a deadline.
func (c *Client) WaitFor(ctx context.Context, jobID string, poll time.Duration) (*contract.Result, error) {
	ticker := time.NewTicker(poll)
	defer ticker.Stop()

	for {
		job, err := c.Get(ctx, jobID)
		if err != nil {
			return nil, err
		}

		if job.Result != nil {
			return job.Result, nil
		}

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("wait for job %s: %w", jobID, ctx.Err())
		case <-ticker.C:
		}
	}
}

// call sends one request and decodes a 2xx answer into out when out is set.
func (c *Client) call(ctx context.Context, method, path string, payload, out any) error {
	var body io.Reader

	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("encode %s %s: %w", method, path, err)
		}

		body = bytes.NewReader(encoded)
	}

	request, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return fmt.Errorf("build %s %s: %w", method, path, err)
	}

	request.Header.Set("Authorization", "Bearer "+c.token)
	request.Header.Set("Content-Type", "application/json")

	response, err := c.http.Do(request)
	if err != nil {
		return fmt.Errorf("%s %s: %w", method, path, err)
	}

	defer func() { _ = response.Body.Close() }()

	if err = decode(response, out); err != nil {
		return fmt.Errorf("%s %s: %w", method, path, err)
	}

	return nil
}
