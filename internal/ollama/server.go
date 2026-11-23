package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/syntopica/worker/internal/config"
)

const (
	residentTimeout = 5 * time.Second
	unloadTimeout   = time.Minute
	launchctlPath   = "/bin/launchctl"
)

// Server is a node's model server.
type Server struct {
	// URL is the server's base URL.
	URL string
	// HTTP sends the requests.
	HTTP *http.Client
	// Run runs launchctl for Restart.
	Run func(ctx context.Context, name string, args ...string) (string, bool)
}

// Resident lists the models the server holds in memory, from /api/ps; false
// when the server does not answer or answers anything else.
func (s *Server) Resident(ctx context.Context) ([]string, bool) {
	raw, ok := s.exchange(ctx, http.MethodGet, "/api/ps", nil, residentTimeout)
	if !ok {
		return nil, false
	}

	var listing struct {
		Models []struct {
			Name *string `json:"name"`
		} `json:"models"`
	}

	if json.Unmarshal(raw, &listing) != nil {
		return nil, false
	}

	names := make([]string, 0, len(listing.Models))

	for _, model := range listing.Models {
		if model.Name == nil {
			return nil, false
		}

		names = append(names, *model.Name)
	}

	return names, true
}

// Unload asks the server to drop a model from memory now: keep_alive 0 on
// /api/generate unloads after the current request.
func (s *Server) Unload(ctx context.Context, model string) bool {
	body := struct {
		Model     string `json:"model"`
		KeepAlive int    `json:"keep_alive"`
	}{Model: model}

	_, ok := s.exchange(ctx, http.MethodPost, "/api/generate", body, unloadTimeout)

	return ok
}

// ProbeQuiet confirms the backend has stopped the previous request (spec 7,
// amendment 2): a one-token request with the pinned options; with
// NUM_PARALLEL=1 it waits its turn.
func (s *Server) ProbeQuiet(ctx context.Context, pin *config.ModelPin, timeout time.Duration) bool {
	body := struct {
		Model     string              `json:"model"`
		Messages  []map[string]string `json:"messages"`
		Stream    bool                `json:"stream"`
		Think     bool                `json:"think"`
		KeepAlive string              `json:"keep_alive"`
		Options   map[string]int      `json:"options"`
	}{
		Model:     pin.Name,
		Messages:  []map[string]string{{"role": "user", "content": "."}},
		KeepAlive: pin.KeepAlive,
		Options:   map[string]int{"num_ctx": pin.NumCtx, "num_predict": 1},
	}

	_, ok := s.exchange(ctx, http.MethodPost, "/api/chat", body, timeout)

	return ok
}

// Restart restarts the model server's LaunchAgent as the last drain recovery
// step: launchctl kickstart -k on the user's GUI domain.
func (s *Server) Restart(ctx context.Context, label string) bool {
	_, ok := s.Run(ctx, launchctlPath, "kickstart", "-k", "gui/"+strconv.Itoa(os.Getuid())+"/"+label)

	return ok
}

// Pause waits, or returns early when the context ends.
func (*Server) Pause(ctx context.Context, wait time.Duration) {
	timer := time.NewTimer(wait)
	defer timer.Stop()

	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}

// exchange sends one request and reads the whole answer; false on a
// transport failure or a status outside 2xx.
func (s *Server) exchange(ctx context.Context, method, path string, payload any, timeout time.Duration) ([]byte, bool) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var body io.Reader

	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, false
		}

		body = bytes.NewReader(data)
	}

	request, err := http.NewRequestWithContext(ctx, method, s.URL+path, body)
	if err != nil {
		return nil, false
	}

	request.Header.Set("Content-Type", "application/json")

	response, err := s.HTTP.Do(request)
	if err != nil {
		return nil, false
	}

	defer func() { _ = response.Body.Close() }()

	raw, err := io.ReadAll(response.Body)
	if err != nil || response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, false
	}

	return raw, true
}
