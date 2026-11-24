package cli_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/syntopica/worker/internal/link"
)

func TestOnDemandSlotsShareSnapshotAndRespectFlags(t *testing.T) {
	var leases atomic.Int32
	var snapshot atomic.Pointer[float64]
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		leases.Add(1)
		var request link.TaskRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
		}
		checkOnDemandLease(t, r, &request, &snapshot)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	newInstance(t, server.Listener.Addr().String())

	code, stdout, stderr := invoke(t, "", "tasks", "--name", "local", "--profile", "fixture", "--snapshot", "--override-pins", "--slots", "3")
	if code != 0 || stdout != "worker: on-demand fixture ran 0 jobs\n" || stderr != "" || leases.Load() != 3 {
		t.Fatalf("tasks = %d, %q, %q, %d leases", code, stdout, stderr, leases.Load())
	}
}

func checkOnDemandLease(t *testing.T, r *http.Request, request *link.TaskRequest, snapshot *atomic.Pointer[float64]) {
	t.Helper()
	if r.URL.Path != "/v1/leases" || r.Header.Get("Authorization") != "Bearer test-local" {
		t.Error("unexpected request path or authorization")
	}
	if request.Profile != "fixture" || !request.OverridePins || request.Node != "local" ||
		request.CreatedBefore == nil || *request.CreatedBefore <= 0 {
		t.Errorf("unexpected on-demand lease request: %+v", request)
		return
	}
	snapshot.CompareAndSwap(nil, request.CreatedBefore)
	if *snapshot.Load() != *request.CreatedBefore {
		t.Error("task slots used different snapshots")
	}
}

func TestEmptyProfileDoesNotStartTheBackgroundLoop(t *testing.T) {
	t.Parallel()
	code, stdout, stderr := invoke(t, "", "tasks", "--name", "local", "--profile=")
	if code != 2 || stdout != "" || !strings.Contains(stderr, "usage: worker tasks") {
		t.Fatalf("empty profile = %d, %q, %q", code, stdout, stderr)
	}
}
