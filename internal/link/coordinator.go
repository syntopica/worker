package link

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/syntopica/worker/internal/contract"
)

const (
	maxPending  = 32
	postTimeout = 30 * time.Second
)

// Coordinator is a node's link to the coordinator.
//
// Transport errors never fail a call: the coordinator may be restarting. A
// completion it could not deliver is held in memory (never on disk: it may
// carry content) and retried before the next lease request; the attempt's
// fencing data makes a late one safe, and lease expiry remains the safety
// net for anything lost with the process.
type Coordinator struct {
	baseURL string
	token   string
	node    string
	http    *http.Client
	pending []held
}

// held is a completion waiting for the coordinator to come back.
type held struct {
	attemptID string
	report    Completion
}

// New returns the link of a node that authenticates with token.
func New(baseURL, token, node string, httpClient *http.Client) *Coordinator {
	return &Coordinator{baseURL: strings.TrimRight(baseURL, "/"), token: token, node: node, http: httpClient}
}

// Lease asks for work: a lease, or nil for no work, a refusal or an
// unreachable coordinator.
func (c *Coordinator) Lease(ctx context.Context, request LeaseRequest) *contract.Lease {
	request.Node = c.node

	return c.lease(ctx, request)
}

// LeaseRemote asks for an escalated inference job for the OpenRouter loop.
func (c *Coordinator) LeaseRemote(ctx context.Context) *contract.Lease {
	return c.lease(ctx, map[string]any{"node": c.node, "kind": "openrouter", "user_active": false, "current_idle_s": 0})
}

// Heartbeat keeps an attempt's lease. It is false only when the attempt was
// fenced out (409): stop, do not complete. An unreachable coordinator keeps
// the attempt.
func (c *Coordinator) Heartbeat(ctx context.Context, attemptID string, generation int64, draining bool) bool {
	body := struct {
		Generation int64 `json:"generation"`
		Draining   bool  `json:"draining"`
	}{generation, draining}

	status, _ := c.post(ctx, "/v1/attempts/"+url.PathEscape(attemptID)+"/heartbeat", body)

	return status != http.StatusConflict
}

// Complete delivers a completion now, or holds it for Flush when the
// coordinator is unreachable.
func (c *Coordinator) Complete(ctx context.Context, attemptID string, generation int64, report *Completion) {
	fenced := *report
	fenced.Generation = generation

	status, _ := c.post(ctx, "/v1/attempts/"+url.PathEscape(attemptID)+"/complete", &fenced)
	if status == 0 && len(c.pending) < maxPending {
		c.pending = append(c.pending, held{attemptID, fenced})
	}
}

// Flush retries the held completions; any answer, even a 409, settles one.
func (c *Coordinator) Flush(ctx context.Context) {
	waiting := c.pending
	c.pending = nil

	for i := range waiting {
		c.Complete(ctx, waiting[i].attemptID, waiting[i].report.Generation, &waiting[i].report)
	}
}

// Pending counts the completions still held.
func (c *Coordinator) Pending() int {
	return len(c.pending)
}

// Report sends the node's host state; best effort.
func (c *Coordinator) Report(ctx context.Context, report *HostReport) {
	c.post(ctx, "/v1/nodes/"+url.PathEscape(c.node)+"/report", report)
}

// lease posts one lease request. Held completions go first; while any is
// still undelivered, the node takes no new work.
func (c *Coordinator) lease(ctx context.Context, request any) *contract.Lease {
	c.Flush(ctx)

	if len(c.pending) > 0 {
		return nil
	}

	status, body := c.post(ctx, "/v1/leases", request)
	if status != http.StatusOK {
		return nil
	}

	var lease contract.Lease
	if json.Unmarshal(body, &lease) != nil {
		return nil
	}

	return &lease
}

// post sends one request and returns its status and body; status 0 means the
// coordinator could not be reached.
func (c *Coordinator) post(ctx context.Context, path string, payload any) (status int, body []byte) {
	data, err := json.Marshal(payload)
	if err != nil {
		return 0, nil
	}

	ctx, cancel := context.WithTimeout(ctx, postTimeout)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return 0, nil
	}

	request.Header.Set("Authorization", "Bearer "+c.token)
	request.Header.Set("Content-Type", "application/json")

	response, err := c.http.Do(request)
	if err != nil {
		return 0, nil
	}

	defer func() { _ = response.Body.Close() }()

	body, err = io.ReadAll(response.Body)
	if err != nil {
		return 0, nil
	}

	return response.StatusCode, body
}
