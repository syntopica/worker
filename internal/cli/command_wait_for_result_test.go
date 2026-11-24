package cli_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestRunReturnsTheJobOutcome(t *testing.T) {
	for _, tc := range []struct {
		name, answer, output, diagnostic, ack string
		code                                  int
	}{
		{
			"JSON wins over text", `{"state":"succeeded","result":{"result_id":"answer","usage":{"tokens_in":1.0},"output":{"json":{"ok":true},"text":"ignored"},"executor":{"provider":"codex","model":"example"}}}`,
			"{\"ok\":true}\n", "result answer executor codex example", `{"result_id":"answer","decline":false}`, 0,
		},
		{
			"control", `{"state":"failed","result":{"result_id":"answer","control":"failed","detail":{"error":"no_eligible_executor"}}}`,
			"", "ended failed no_eligible_executor", `{"result_id":"answer","decline":false}`, 1,
		},
		{
			"declined split", `{"state":"awaiting_split","result":{"result_id":"answer","control":"split_requested"}}`,
			"", "still awaiting_split", `{"result_id":"answer","decline":true}`, 4,
		},
		{"parked", `{"state":"queued","result":null,"cooling_until":12345}`, "", "parked behind a quota wall", "", 3},
		{"pending", `{"state":"queued","result":null}`, "", "still queued after 0s", "", 4},
		{"already collected", `{"state":"succeeded","result":null}`, "", "with nothing to collect", "", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var acks atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				serveRunOutcome(t, w, r, tc.answer, tc.ack, &acks)
			}))
			defer server.Close()
			newInstance(t, server.Listener.Addr().String())

			code, stdout, stderr := invoke(t, `{"kind":"task"}`, "run", "-", "--producer", "producer", "--wait", "0")
			if code != tc.code || stdout != tc.output || !strings.Contains(stderr, tc.diagnostic) {
				t.Fatalf("run = %d, %q, %q", code, stdout, stderr)
			}
			if got := acks.Load(); got > 1 || (got == 0) != (tc.ack == "") {
				t.Errorf("unexpected ACK count %d", got)
			}
		})
	}
}

func serveRunOutcome(t *testing.T, w http.ResponseWriter, r *http.Request, answer, ack string, acks *atomic.Int32) {
	t.Helper()
	switch r.URL.Path {
	case "/v1/jobs":
		assertRequest(t, r, http.MethodPost, "/v1/jobs", `{"kind":"task"}`)
		reply(t, w, `{"id":"job","created":true}`)
	case "/v1/jobs/job":
		assertRequest(t, r, http.MethodGet, "/v1/jobs/job", "")
		reply(t, w, answer)
	case "/v1/jobs/job/ack":
		acks.Add(1)
		assertRequest(t, r, http.MethodPost, "/v1/jobs/job/ack", ack)
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}
