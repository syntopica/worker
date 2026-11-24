package coordinator_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/syntopica/worker/internal/auth"
	"github.com/syntopica/worker/internal/client"
	"github.com/syntopica/worker/internal/coordinator"
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
	node, err := auth.AddPrincipal(state, "node", "local")
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
	return testServer{config: path, url: "http://" + server.Addr().String(), producer: producer, node: node}
}

const pythonRoundTrip = `
import http.client
import json
import sys
import time
from dataclasses import replace
from pathlib import Path

from urllib.parse import urlsplit

from tests.node.fake_ollama import FakeOllama
from worker.client.worker_client import WorkerClient
from worker.config.load_worker_config import load_worker_config
from worker.node.coordinator_link import CoordinatorLink
from worker.node.host_state import HostState
from worker.node.run_node import run_node

base, producer_token, node_token, config_path = json.load(sys.stdin)
producer = WorkerClient(base, producer_token)
node = CoordinatorLink(base, node_token, "local")
job = {
    "contract": 1,
    "kind": "inference",
    "privacy": "mail",
    "queue": "work",
    "idempotency_key": "python-round-trip",
    "requirements": {"models": ["model"]},
    "input": {"messages": [{"role": "user", "content": "fixture input"}]},
}
submitted = producer.submit(job)
lease = node.lease(None, False, 10, 900)
assert lease and lease["job_id"] == submitted["id"]
attempt, generation = lease["attempt_id"], lease["generation"]
assert node.heartbeat(attempt, generation, False)
node.complete(attempt, generation, {
    "outcome": "succeeded",
    "output": {"text": "fixture output", "json": None},
    "executor": {"provider": "ollama", "model": "model", "node": "local"},
    "usage": {"tokens_in": 2, "tokens_out": 1},
})
assert not node.heartbeat(attempt, generation, False)
[result] = producer.results("work")
assert result["output"] == {"text": "fixture output", "json": None}
assert result["executor"] == {"provider": "ollama", "model": "model", "node": "local"}
producer.ack(submitted["id"], result["result_id"])
assert not producer.results("work")

# Run the unchanged inference loop against the real Go coordinator.
fake = FakeOllama()
fake.loaded = ["model"]
try:
    cfg = load_worker_config(Path(config_path))
    cfg = replace(cfg, nodes={"local": replace(cfg.nodes["local"], ollama_url=fake.url)})
    producer.submit({**job, "idempotency_key": "python-node"})
    offset = [0.0]
    class StopLoop(Exception):
        pass
    def sleep(seconds):
        if seconds < 10:
            time.sleep(0.05)
            return
        offset[0] += 200.0
        if offset[0] >= 400.0:
            raise StopLoop
    try:
        run_node(cfg, "local", node, sample=lambda: HostState(900, True, "normal"),
                 sleep=sleep, clock=lambda: time.time() + offset[0])
    except StopLoop:
        pass
    [result] = producer.results("work")
    assert result["output"]["json"] == {"label": "x"}
    assert fake.bodies
finally:
    fake.close()

# A producer writes the entire body before it reads the error response.
conn = http.client.HTTPConnection(urlsplit(base).netloc, timeout=5)
try:
    conn.request("POST", "/v1/jobs", b"x" * (8 * 1024 * 1024))
    refused = conn.getresponse()
    assert refused.status == 413
    assert json.loads(refused.read()) == {"error": "payload_too_large"}
finally:
    conn.close()
`

func TestPythonNodeRoundTrip(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("Python 3 is needed to check cross-language compatibility")
	}
	f := startServer(t)
	data, err := json.Marshal([4]string{f.url, f.producer, f.node, f.config})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "python3", "-c", pythonRoundTrip)
	cmd.Dir = "../.."
	cmd.Stdin = strings.NewReader(string(data))
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Python clients failed: %v\n%s", err, output)
	}
}
