package coordinator_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/syntopica/worker/internal/auth"
	"github.com/syntopica/worker/internal/client"
	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/contract"
	"github.com/syntopica/worker/internal/coordinator"
	"github.com/syntopica/worker/internal/host"
	"github.com/syntopica/worker/internal/link"
	"github.com/syntopica/worker/internal/node"
	"github.com/syntopica/worker/internal/ollama"
	"github.com/syntopica/worker/internal/ollamatest"
)

const fixtureConfig = `{
  "listen":"127.0.0.1:0",
  "models":{"model":{"num_ctx":4096,"keep_alive":"5m","cold_gb":2,"warm_gb":1}},
  "queues":{"work":{"run_when":"idle","max_outstanding":3}},
  "nodes":{"local":{"trust":"owner","memory_budget_gb":10,"ollama_url":"http://127.0.0.1:1","ollama_launchd_label":"test.ollama"}},
  "producers":{"producer":["work"]}
}`

func TestServerUsesTheReloadedConfig(t *testing.T) {
	t.Parallel()
	f := startServer(t)
	producer := client.New(f.url, f.producer, nil)
	if err := os.WriteFile(f.config, []byte(strings.Replace(fixtureConfig, "127.0.0.1", "0.0.0.0", 1)), 0o600); err != nil {
		t.Fatal(err)
	}
	// The invalid reload leaves the initial producer grant in force.
	if _, err := producer.Results(t.Context(), client.ResultsQuery{Queue: "work"}); err != nil {
		t.Fatal(err)
	}
	changed := strings.Replace(fixtureConfig, `{"producer":["work"]}`, "{}", 1)
	if err := os.WriteFile(f.config, []byte(changed), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := producer.Results(t.Context(), client.ResultsQuery{Queue: "work"})
	if refusal, ok := errors.AsType[*client.Error](err); !ok || refusal.Status != http.StatusForbidden || refusal.Code != "queue_not_granted" {
		t.Fatalf("Results after removing the grant = %v; want queue_not_granted", err)
	}
}

func TestServerRefusesNonLoopback(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(strings.Replace(fixtureConfig, "127.0.0.1", "0.0.0.0", 1)), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := coordinator.Listen(t.Context(), path, filepath.Join(dir, "state"), t.Output()); err == nil {
		t.Fatal("non-loopback listen address was accepted")
	}
}

type testServer struct {
	config, url, producer, node string
}

func startServer(t *testing.T) testServer {
	t.Helper()
	dir := t.TempDir()
	path, state := filepath.Join(dir, "config.json"), filepath.Join(dir, "state")
	if err := os.WriteFile(path, []byte(fixtureConfig), 0o600); err != nil {
		t.Fatal(err)
	}
	producer, err := auth.AddPrincipal(state, "producer", "producer")
	if err != nil {
		t.Fatal(err)
	}
	nodeToken, err := auth.AddPrincipal(state, "node", "local")
	if err != nil {
		t.Fatal(err)
	}
	server, err := coordinator.Listen(t.Context(), path, state, t.Output())
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- server.Run(t.Context()) }()
	t.Cleanup(func() {
		if err := <-done; err != nil {
			t.Error(err)
		}
	})
	return testServer{config: path, url: "http://" + server.Addr().String(), producer: producer, node: nodeToken}
}

func TestCoordinatorProducerAndNodeRoundTrip(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	f := startServer(t)
	producer := client.New(f.url, f.producer, nil)
	submitted := submitInference(t, producer, "fenced-round-trip")
	nodeLink := link.New(f.url, f.node, "local", http.DefaultClient)
	request := link.LeaseRequest{FreeGB: 10, CurrentIdleS: 900}
	lease := nodeLink.Lease(ctx, request)
	if lease == nil || lease.JobID != submitted {
		t.Fatalf("lease = %+v, want %s", lease, submitted)
	}
	if !nodeLink.Heartbeat(ctx, lease.AttemptID, lease.Generation, false) {
		t.Fatal("current attempt was fenced")
	}
	completion := &link.Completion{
		Outcome:  "succeeded",
		Output:   map[string]any{"text": "fixture output", "json": nil},
		Executor: contract.Executor{Provider: "ollama", Model: "model", Node: "local"},
		Usage:    map[string]any{"tokens_in": 2, "tokens_out": 1},
	}
	nodeLink.Complete(ctx, lease.AttemptID, lease.Generation+1, completion)
	job, err := producer.Get(ctx, lease.JobID)
	if err != nil || job.Result != nil {
		t.Fatalf("stale completion created a result: %v", err)
	}
	nodeLink.Complete(ctx, lease.AttemptID, lease.Generation, completion)
	if nodeLink.Heartbeat(ctx, lease.AttemptID, lease.Generation, false) {
		t.Fatal("settled attempt was not fenced")
	}
	assertCompletionRoundTrip(t, producer, lease.JobID, completion.Executor)
}

func assertCompletionRoundTrip(t *testing.T, producer *client.Client, id string, executor contract.Executor) {
	t.Helper()
	result := inferenceResult(t, producer)
	if string(result.Output) != `{"json":null,"text":"fixture output"}` ||
		result.Executor == nil || *result.Executor != executor {
		t.Fatal("completion did not round trip")
	}
	if err := producer.Ack(t.Context(), id, contract.Ack{ResultID: result.ResultID}); err != nil {
		t.Fatal(err)
	}
	rows, err := producer.Results(t.Context(), client.ResultsQuery{Queue: "work"})
	if err != nil || len(rows) != 0 {
		t.Fatalf("results after ack = %d, %v", len(rows), err)
	}
}

func TestCoordinatorRoundTripWithGoInferenceLoop(t *testing.T) {
	t.Parallel()
	f := startServer(t)
	producer := client.New(f.url, f.producer, nil)
	submitInference(t, producer, "go-node")
	fake := ollamatest.New(t)
	fake.Set(func(s *ollamatest.Fake) { s.Loaded = []string{"model"} })
	cfg, err := config.Load(f.config)
	if err != nil {
		t.Fatal(err)
	}
	policy := cfg.Nodes["local"]
	policy.OllamaURL = fake.URL
	cfg.Nodes["local"] = policy
	runInferenceLoop(t, f, cfg)

	result := inferenceResult(t, producer)
	var answer struct {
		JSON map[string]string `json:"json"`
	}
	if err := json.Unmarshal(result.Output, &answer); err != nil || answer.JSON["label"] != "x" || len(fake.Bodies()) == 0 {
		t.Fatalf("fake Ollama result was not collected: %v", err)
	}
}

func TestCoordinatorReturnsARefusalAfterOversizeWrite(t *testing.T) {
	t.Parallel()
	f := startServer(t)
	const size = 8 * 1024 * 1024
	body := strings.NewReader(strings.Repeat("x", size))
	request, err := http.NewRequestWithContext(t.Context(), http.MethodPost, f.url+"/v1/jobs", body)
	if err != nil {
		t.Fatal(err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = response.Body.Close() }()
	var refusal contract.Refusal
	if err := json.NewDecoder(response.Body).Decode(&refusal); err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusRequestEntityTooLarge || refusal.Error != "payload_too_large" {
		t.Fatalf("oversize request = %d %s", response.StatusCode, refusal.Error)
	}
}

func submitInference(t *testing.T, producer *client.Client, key string) string {
	t.Helper()
	const document = `{
  "contract":1, "kind":"inference", "privacy":"mail", "queue":"work", "idempotency_key":%q,
  "requirements":{"models":["model"]}, "input":{"messages":[{"role":"user","content":"fixture input"}]}
}`
	submitted, err := producer.SubmitDocument(t.Context(), json.RawMessage(fmt.Sprintf(document, key)))
	if err != nil {
		t.Fatal(err)
	}
	return submitted.ID
}

func inferenceResult(t *testing.T, producer *client.Client) *contract.Result {
	t.Helper()
	rows, err := producer.Results(t.Context(), client.ResultsQuery{Queue: "work"})
	if err != nil || len(rows) != 1 {
		t.Fatalf("results = %d, %v", len(rows), err)
	}
	return &rows[0]
}

func runInferenceLoop(t *testing.T, f testServer, cfg *config.Config) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	idle, onAC := 15*time.Minute, true
	var rests int
	var now time.Time
	run := &node.Node{
		Name: "local", Config: func() *config.Config { return cfg },
		Link: link.New(f.url, f.node, "local", http.DefaultClient),
		Backend: func(url string) ollama.Backend {
			return &ollama.Server{URL: url, HTTP: http.DefaultClient}
		},
		HTTP: http.DefaultClient, Log: t.Output(),
		Sample: func(context.Context) host.State {
			return host.State{Idle: &idle, OnAC: &onAC, Pressure: host.PressureNormal}
		},
		Now: func() time.Time { return now },
		Sleep: func(ctx context.Context, wait time.Duration) {
			if wait >= 30*time.Second {
				now = now.Add(3 * time.Minute)
				rests++
				if rests == 2 {
					cancel()
				}
				return
			}
			timer := time.NewTimer(wait)
			defer timer.Stop()
			select {
			case <-ctx.Done():
			case <-timer.C:
			}
		},
	}
	if err := run.Run(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("inference loop = %v", err)
	}
}
