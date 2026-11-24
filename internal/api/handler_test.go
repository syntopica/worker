package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/syntopica/worker/internal/api"
	"github.com/syntopica/worker/internal/auth"
	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/store"
)

const submission = `{"contract":1,"kind":"inference","privacy":"mail","queue":"work","idempotency_key":"example","requirements":{"models":["model"]},"input":{"messages":[{"role":"user","content":"fixture input"}]}}`

type fixture struct {
	handler http.Handler
	tokens  map[string]string
	cfg     *config.Config
}

func openAPI(t *testing.T) fixture {
	t.Helper()
	dir := t.TempDir()
	if err := store.Migrate(t.Context(), dir); err != nil {
		t.Fatal(err)
	}
	s, err := store.Open(t.Context(), dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := s.Close(); err != nil {
			t.Error(err)
		}
	})
	tokens := make(map[string]string)
	for name, kind := range map[string]string{"producer": "producer", "other": "producer", "local": "node", "guest": "node", "admin": "admin"} {
		token, err := auth.AddPrincipal(dir, kind, name)
		if err != nil {
			t.Fatal(err)
		}
		tokens[name] = token
	}
	cfg := &config.Config{
		MaxPayloadBytes: 1024,
		Producers:       map[string][]string{"producer": {"work"}},
		Models:          map[string]config.ModelPin{"model": {ColdGB: 2}},
		Nodes: map[string]config.NodePolicy{
			"local": {Trust: "owner", MemoryBudgetGB: 10},
			"guest": {Trust: "guest", MemoryBudgetGB: 10},
		},
		Queues:  map[string]config.QueuePolicy{"work": {RunWhen: "idle", Weight: 1, MaxOutstanding: 10}},
		Privacy: map[string][]string{"mail": {"ollama"}},
		Trust:   map[string][]string{"mail": {"owner"}},
	}
	return fixture{handler: api.NewHandler(s, dir, func() *config.Config { return cfg }, t.Output()), tokens: tokens, cfg: cfg}
}

func (f fixture) request(t *testing.T, method, path, user, body string, status int) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequestWithContext(t.Context(), method, path, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+f.tokens[user])
	return f.send(t, req, status)
}

func (f fixture) send(t *testing.T, req *http.Request, status int) *httptest.ResponseRecorder {
	t.Helper()
	reply := httptest.NewRecorder()
	f.handler.ServeHTTP(reply, req)
	if reply.Code != status {
		t.Fatalf("%s %s returned %d with %s; want %d", req.Method, req.URL, reply.Code, reply.Body, status)
	}
	return reply
}

func response[T any](t *testing.T, reply *httptest.ResponseRecorder) T {
	t.Helper()
	var data T
	if err := json.Unmarshal(reply.Body.Bytes(), &data); err != nil {
		t.Fatal(err)
	}
	return data
}

func wantJSON(t *testing.T, reply *httptest.ResponseRecorder, fragment string) {
	t.Helper()
	if !strings.Contains(reply.Body.String(), fragment) {
		t.Fatalf("got %s; want %s", reply.Body, fragment)
	}
}

func TestInternalRefusalDoesNotExposePanicContent(t *testing.T) {
	t.Parallel()
	var log bytes.Buffer
	handler := api.NewHandler(nil, t.TempDir(), func() *config.Config {
		panic("fixture content must stay private")
	}, &log)
	reply := httptest.NewRecorder()
	handler.ServeHTTP(reply, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/v1/status", nil))
	if reply.Code != http.StatusInternalServerError || reply.Body.String() != `{"error": "internal"}` {
		t.Fatalf("unexpected 500 response: %d %s", reply.Code, reply.Body)
	}
	if log.String() != "worker: request failed\n" {
		t.Fatalf("unexpected diagnostic: %q", log.String())
	}
}
