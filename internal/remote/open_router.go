package remote

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"
)

const (
	callTimeout     = 600 * time.Second
	headroomTimeout = 10 * time.Second
)

// OpenRouter is the OpenRouter API under one key.
type OpenRouter struct {
	// API is the base URL.
	API string
	// Key authenticates every request.
	Key string
	// HTTP sends the requests.
	HTTP *http.Client
}

// Headroom is the free-model requests left today, as OpenRouter reports them
// (spec 8, quotas); false when unknown, which means no headroom.
func (o *OpenRouter) Headroom(ctx context.Context) (left int, known bool) {
	ctx, cancel := context.WithTimeout(ctx, headroomTimeout)
	defer cancel()

	status, raw := o.send(ctx, http.MethodGet, "/key", nil)
	if status < http.StatusOK || status >= http.StatusMultipleChoices {
		return 0, false
	}

	var answer struct {
		Data struct {
			Free struct {
				Remaining *float64 `json:"remaining"`
			} `json:"free_model_daily_requests"`
		} `json:"data"`
	}

	if json.Unmarshal(raw, &answer) != nil || answer.Data.Free.Remaining == nil {
		return 0, false
	}

	return int(*answer.Data.Free.Remaining), true
}

// Post sends one chat completion. The codes are allowlisted: rate_limited
// for a 429, http_<status> for another refusal, transport_error when
// unreachable, bad_response for a body that is not a completion.
func (o *OpenRouter) Post(ctx context.Context, body map[string]any) (answer *Answer, code string) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, string(badResponse)
	}

	ctx, cancel := context.WithTimeout(ctx, callTimeout)
	defer cancel()

	status, raw := o.send(ctx, http.MethodPost, "/chat/completions", data)

	switch {
	case status == 0:
		return nil, string(transportError)
	case status == http.StatusTooManyRequests:
		return nil, string(rateLimited)
	case status < http.StatusOK || status >= http.StatusMultipleChoices:
		return nil, "http_" + strconv.Itoa(status)
	}

	return decodeAnswer(raw)
}

// send makes one request and returns its status and body; status 0 means
// OpenRouter could not be reached.
func (o *OpenRouter) send(ctx context.Context, method, path string, data []byte) (status int, body []byte) {
	request, err := http.NewRequestWithContext(ctx, method, o.API+path, bytes.NewReader(data))
	if err != nil {
		return 0, nil
	}

	request.Header.Set("Authorization", "Bearer "+o.Key)

	if data != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	response, err := o.HTTP.Do(request)
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

// decodeAnswer reads a chat-completions answer; one that carries an error or
// no choice is not a completion.
func decodeAnswer(raw []byte) (answer *Answer, code string) {
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil {
		return nil, string(badResponse)
	}

	var completion struct {
		Model   *string `json:"model"`
		Choices []struct {
			Message struct {
				Content *string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     float64 `json:"prompt_tokens"`
			CompletionTokens float64 `json:"completion_tokens"`
			Cost             float64 `json:"cost"`
		} `json:"usage"`
	}

	_, refused := fields["error"]
	if refused || json.Unmarshal(raw, &completion) != nil || len(completion.Choices) == 0 {
		return nil, string(badResponse)
	}

	answer = &Answer{
		TokensIn:  int(completion.Usage.PromptTokens),
		TokensOut: int(completion.Usage.CompletionTokens),
		CostUSD:   completion.Usage.Cost,
	}

	if completion.Model != nil {
		answer.Model = *completion.Model
	}

	if content := completion.Choices[0].Message.Content; content != nil {
		answer.Content = *content
	}

	return answer, ""
}
