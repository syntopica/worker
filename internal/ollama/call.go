package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"
)

// Phases of a call, for the cancel-to-quiet measurement, and the error codes
// of one that did not answer.
const (
	phasePrefill       = "prefill"
	phaseGeneration    = "generation"
	codeTransportError = "transport_error"
	codeCancelled      = "cancelled"
)

// Call is one /api/chat request in the background, cancellable by closing
// its connection. Closing is only a hint to the server; Drain confirms quiet.
type Call struct {
	stop       context.CancelFunc
	finished   chan struct{}
	cancelled  atomic.Bool
	generating atomic.Bool
	answer     *Answer
	code       string
}

// StartCall sends the request in the background and returns immediately.
// idle is the longest silence tolerated from the server, between chunks too.
func StartCall(ctx context.Context, client *http.Client, url string, body any, idle time.Duration) *Call {
	ctx, stop := context.WithCancel(ctx)
	call := &Call{stop: stop, finished: make(chan struct{})}

	go func() {
		defer close(call.finished)
		defer stop()

		call.answer, call.code = call.run(ctx, client, url, body, idle)
	}()

	return call
}

// Phase is "generation" once the first chunk arrived, else "prefill".
func (c *Call) Phase() string {
	if c.generating.Load() {
		return phaseGeneration
	}

	return phasePrefill
}

// Done reports whether the request has finished, for any reason.
func (c *Call) Done() bool {
	select {
	case <-c.finished:
		return true
	default:
		return false
	}
}

// Finished is closed once the request has finished, for any reason.
func (c *Call) Finished() <-chan struct{} {
	return c.finished
}

// Cancel closes the connection so the blocked read fails.
func (c *Call) Cancel() {
	c.cancelled.Store(true)
	c.stop()
}

// Outcome is the answer, or nil and an error code. It waits for the request
// to finish; a cancelled call is "cancelled" whatever it had read.
func (c *Call) Outcome() (answer *Answer, code string) {
	<-c.finished

	if c.cancelled.Load() {
		return nil, codeCancelled
	}

	return c.answer, c.code
}

// run performs the request; a transport failure is transport_error.
func (c *Call) run(ctx context.Context, client *http.Client, url string, body any, idle time.Duration) (answer *Answer, code string) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, codeTransportError
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url+"/api/chat", bytes.NewReader(data))
	if err != nil {
		return nil, codeTransportError
	}

	request.Header.Set("Content-Type", "application/json")

	silence := time.AfterFunc(idle, c.stop)
	defer silence.Stop()

	response, err := client.Do(request)
	if err != nil {
		return nil, codeTransportError
	}

	defer func() { _ = response.Body.Close() }()

	if response.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, response.Body)

		return nil, "http_" + strconv.Itoa(response.StatusCode)
	}

	answer, code = readChatStream(response.Body, func() {
		c.generating.Store(true)
		silence.Reset(idle)
	})
	if answer == nil && ctx.Err() != nil {
		// The stream was cut, not malformed.
		return nil, codeTransportError
	}

	return answer, code
}
