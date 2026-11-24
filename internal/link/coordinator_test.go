package link_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"

	"github.com/syntopica/worker/internal/contract"
	"github.com/syntopica/worker/internal/link"
)

// recorder is a coordinator that answers every request with one status and
// body and records what it received.
type recorder struct {
	mu     sync.Mutex
	status int
	body   string
	seen   []string
	auth   string
	down   bool
}

func (r *recorder) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	raw, _ := io.ReadAll(request.Body)

	r.mu.Lock()
	r.seen = append(r.seen, request.URL.Path+" "+string(raw))
	r.auth = request.Header.Get("Authorization")
	status, body, down := r.status, r.body, r.down
	r.mu.Unlock()

	if hijacker, ok := w.(http.Hijacker); ok && down {
		if conn, _, err := hijacker.Hijack(); err == nil {
			_ = conn.Close()
		}

		return
	}

	w.WriteHeader(status)
	_, _ = w.Write([]byte(body))
}

// requests are the path and body of every request received so far.
func (r *recorder) requests() []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	return append([]string(nil), r.seen...)
}

// authorization is the last Authorization header received.
func (r *recorder) authorization() string {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.auth
}

// setDown makes the coordinator drop every connection, or answer again.
func (r *recorder) setDown(down bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.down = down
}

// serve starts a recording coordinator and a link to it.
func serve(t *testing.T, status int, body string) (*recorder, *link.Coordinator) {
	t.Helper()

	coordinator := &recorder{status: status, body: body}
	server := httptest.NewServer(coordinator)
	t.Cleanup(server.Close)

	return coordinator, link.New(server.URL+"/", "node-token", "node-1", server.Client())
}

// unreachable is a link to a coordinator that is down.
func unreachable() *link.Coordinator {
	return link.New("http://127.0.0.1:1", "node-token", "node-1", http.DefaultClient)
}

func failed() *link.Completion {
	return &link.Completion{Outcome: "failed", Executor: contract.Executor{Node: "node-1", Provider: "ollama", Model: "model-a"}}
}

func TestCoordinatorUnreachableNeverFails(t *testing.T) {
	t.Parallel()

	down := unreachable()

	if down.Lease(t.Context(), link.LeaseRequest{FreeGB: 44, CurrentIdleS: 900}) != nil {
		t.Error("an unreachable coordinator leased")
	}

	if !down.Heartbeat(t.Context(), "a", 1, false) {
		t.Error("an unreachable coordinator fenced the attempt")
	}

	down.Report(t.Context(), &link.HostReport{Pressure: "normal"})

	if down.Pending() != 0 {
		t.Errorf("pending = %d, want 0", down.Pending())
	}
}

func TestCoordinatorHeartbeatIsFalseOnlyWhenFenced(t *testing.T) {
	t.Parallel()

	for status, alive := range map[int]bool{200: true, 409: false, 500: true, 401: true} {
		coordinator, node := serve(t, status, "{}")

		if node.Heartbeat(t.Context(), "a", 3, true) != alive {
			t.Errorf("status %d: alive = %v, want %v", status, !alive, alive)
		}

		if want := `/v1/attempts/a/heartbeat {"generation":3,"draining":true}`; coordinator.requests()[0] != want {
			t.Errorf("status %d: seen = %v", status, coordinator.requests())
		}
	}
}

func TestCoordinatorLease(t *testing.T) {
	t.Parallel()

	lease := `{"job_id":"j1","attempt_id":"a1","generation":2,"model":"model-a","input":{"messages":[]},` +
		`"run_when":"idle","ttl_s":120,"kind":"inference","queue":"q","privacy":"public","pinned":false}`
	coordinator, node := serve(t, 200, lease)
	resident := "model-a"

	got := node.Lease(t.Context(), link.LeaseRequest{ResidentModel: &resident, UserActive: true, FreeGB: 12.5, CurrentIdleS: 3})
	if got == nil || got.AttemptID != "a1" || got.Generation != 2 || got.RunWhen != "idle" {
		t.Fatalf("lease = %+v", got)
	}

	want := `/v1/leases {"node":"node-1","resident_model":"model-a","user_active":true,"free_gb":12.5,"current_idle_s":3}`
	if coordinator.requests()[0] != want || coordinator.authorization() != "Bearer node-token" {
		t.Errorf("seen = %v, auth = %q", coordinator.requests(), coordinator.authorization())
	}

	for status, body := range map[int]string{204: "", 403: `{"error":"forbidden"}`, 200: "not a lease"} {
		if _, refusing := serve(t, status, body); refusing.Lease(t.Context(), link.LeaseRequest{}) != nil {
			t.Errorf("status %d body %q leased", status, body)
		}
	}
}

func TestCoordinatorCompleteAndReportBodies(t *testing.T) {
	t.Parallel()

	coordinator, node := serve(t, 200, `{"ok":true}`)
	code := "node_error"
	report := failed()
	report.ErrorCode, report.Usage, report.WallS = &code, map[string]any{}, 1.5

	node.Complete(t.Context(), "a1", 4, report)
	node.Report(t.Context(), &link.HostReport{Pressure: "normal", Resident: []string{}, Unexpected: []string{}})

	want := []string{
		`/v1/attempts/a1/complete {"outcome":"failed","output":null,"usage":{},` +
			`"executor":{"node":"node-1","provider":"ollama","model":"model-a"},"error_code":"node_error","wall_s":1.5,"generation":4}`,
		`/v1/nodes/node-1/report {"reason":null,"idle_s":null,"on_ac":null,"pressure":"normal",` +
			`"resident":[],"unexpected":[],"loads_1h":0,"unloads_1h":0}`,
	}
	if !reflect.DeepEqual(coordinator.requests(), want) {
		t.Errorf("seen = %v\nwant   %v", coordinator.requests(), want)
	}

	if node.Pending() != 0 {
		t.Errorf("pending = %d after a delivered completion", node.Pending())
	}

	node.Complete(t.Context(), "a1", 4, &link.Completion{Output: map[string]any{"bad": json.Number("x")}, Usage: nil})
}

func TestCoordinatorHoldsAnUndeliveredCompletion(t *testing.T) {
	t.Parallel()

	down := unreachable()
	down.Complete(t.Context(), "a", 1, failed())

	if down.Pending() != 1 {
		t.Fatalf("pending = %d, want 1", down.Pending())
	}

	if down.Lease(t.Context(), link.LeaseRequest{}) != nil || down.Pending() != 1 {
		t.Errorf("a held completion did not block the lease: pending = %d", down.Pending())
	}

	for range 40 {
		down.Complete(t.Context(), "b", 1, failed())
	}

	if down.Pending() != 32 {
		t.Errorf("pending = %d, want the cap of 32", down.Pending())
	}
}

func TestCoordinatorDeliversAHeldCompletionBeforeTheNextLease(t *testing.T) {
	t.Parallel()

	coordinator, node := serve(t, 409, "{}")
	coordinator.setDown(true)

	node.Complete(t.Context(), "a", 7, failed())

	if node.Pending() != 1 || len(coordinator.requests()) != 1 {
		t.Fatalf("pending = %d, seen = %v", node.Pending(), coordinator.requests())
	}

	coordinator.setDown(false)

	if node.Lease(t.Context(), link.LeaseRequest{}) != nil {
		t.Error("a 409 leased")
	}

	if node.Pending() != 0 {
		t.Errorf("pending = %d: a 409 did not settle the held completion", node.Pending())
	}

	held := `/v1/attempts/a/complete {"outcome":"failed","output":null,"usage":null,` +
		`"executor":{"node":"node-1","provider":"ollama","model":"model-a"},"error_code":null,"wall_s":0,"generation":7}`
	want := []string{held, held, `/v1/leases {"node":"node-1","resident_model":null,"user_active":false,"free_gb":0,"current_idle_s":0}`}

	if !reflect.DeepEqual(coordinator.requests(), want) {
		t.Errorf("seen = %v\nwant   %v", coordinator.requests(), want)
	}
}

func TestCoordinatorLeaseRemoteAsksForAnEscalatedJob(t *testing.T) {
	t.Parallel()

	coordinator, node := serve(t, http.StatusOK, `{"job_id": "j", "attempt_id": "a", "model": "vendor/model:free"}`)

	lease := node.LeaseRemote(t.Context())
	if lease == nil || lease.Model != "vendor/model:free" {
		t.Fatalf("lease %+v", lease)
	}

	want := `/v1/leases {"current_idle_s":0,"kind":"openrouter","node":"node-1","user_active":false}`
	if got := coordinator.requests(); len(got) != 1 || got[0] != want {
		t.Errorf("requests %v", got)
	}
}
