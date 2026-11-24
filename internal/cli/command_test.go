package cli_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/syntopica/worker/internal/cli"
)

func TestSubmitPreservesTheJobDocument(t *testing.T) {
	const document = `{"kind":"inference","requirements":{"exclude_providers":["ollama"]}}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertRequest(t, r, http.MethodPost, "/v1/jobs", document)
		reply(t, w, `{"id":"job","created":true}`)
	}))
	defer server.Close()
	newInstance(t, server.Listener.Addr().String())
	path := writeJob(t, document)

	code, stdout, stderr := invoke(t, "", "submit", path, "--producer", "producer")
	if code != 0 || strings.TrimSpace(stdout) != `{"id":"job","created":true}` || stderr != "" {
		t.Fatalf("submit = %d, %q, %q", code, stdout, stderr)
	}
}

func TestWorkerUsageDoesNotQuoteItsArguments(t *testing.T) {
	t.Parallel()
	code, stdout, stderr := invoke(t, "", "run", "--wait", "private payload")
	if code != 2 || stdout != "" || !strings.Contains(stderr, "usage: worker run") || strings.Contains(stderr, "private payload") {
		t.Fatalf("invalid flags = %d, %q, %q", code, stdout, stderr)
	}
}

func newInstance(t *testing.T, address string) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("SYNTOPICA_DATA", dir)
	root := filepath.Join(dir, "worker")
	tokens := filepath.Join(root, "state", "tokens")
	if err := os.MkdirAll(tokens, 0o700); err != nil {
		t.Fatal(err)
	}
	cfg := fmt.Sprintf(`{"listen":%q,"models":{},"queues":{},"producers":{},"nodes":{"local":{"trust":"owner","memory_budget_gb":10,"ollama_url":"","ollama_launchd_label":"","task_slots":2}}}`, address)
	if err := os.WriteFile(filepath.Join(root, "config.json"), []byte(cfg), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"producer", "local", "admin"} {
		if err := os.WriteFile(filepath.Join(tokens, name+".token"), []byte("test-"+name), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func writeJob(t *testing.T, raw string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "job.json")
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func invoke(t *testing.T, input string, args ...string) (code int, out, diagnostics string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code = cli.Run(t.Context(), args, strings.NewReader(input), &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

func assertRequest(t *testing.T, r *http.Request, method, path, body string) {
	t.Helper()
	raw, err := io.ReadAll(r.Body)
	if err != nil || r.Method != method || r.URL.Path != path || string(raw) != body {
		t.Errorf("request %s %s with %q: %v", r.Method, r.URL.Path, raw, err)
	}
}

func reply(t *testing.T, w http.ResponseWriter, body string) {
	t.Helper()
	if _, err := io.WriteString(w, body); err != nil {
		t.Error(err)
	}
}
