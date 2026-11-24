package jobs_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/contract"
	"github.com/syntopica/worker/internal/jobs"
	"github.com/syntopica/worker/internal/policy"
)

func leaseAt(t *testing.T, req *policy.LeaseRequest, s jobTestStore, now float64) *contract.Lease {
	t.Helper()
	lease, err := jobs.LeaseJob(t.Context(), s.store, s.cfg, req, now)
	if err != nil || lease == nil {
		t.Fatalf("LeaseJob() = %v, %v; want a lease", lease, err)
	}
	return lease
}

func TestCompleteFencesBeforeWriting(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	s, cfg := openJobs(t), jobConfig()
	id := submit(t, s, cfg, document)
	lease := leaseAt(t, &policy.LeaseRequest{Node: "local"}, jobTestStore{s, cfg}, 1)
	report := &jobs.CompletionReport{
		Outcome: "succeeded", Output: map[string]any{"text": "answer"},
		Executor: map[string]any{"node": "local", "provider": "ollama", "model": "model"},
	}
	_, err := jobs.CompleteAttempt(ctx, s, cfg, lease.AttemptID, lease.Generation, "other", report, 2)
	requireRefusal(t, err, http.StatusConflict, "stale_attempt")
	state, err := jobs.CompleteAttempt(ctx, s, cfg, lease.AttemptID, lease.Generation, "local", report, 2)
	if err != nil || state != "succeeded" {
		t.Fatalf("CompleteAttempt() = %q, %v; want succeeded", state, err)
	}
	_, err = jobs.CompleteAttempt(ctx, s, cfg, lease.AttemptID, lease.Generation, "local", report, 2)
	requireRefusal(t, err, http.StatusConflict, "stale_attempt")
	var output string
	err = s.QueryRowContext(ctx, "SELECT p.body FROM results r JOIN p.outputs p USING (result_id) WHERE r.job_id = ?", id).Scan(&output)
	if err != nil || output != `{"text":"answer"}` {
		t.Fatalf("stored payload = %q, %v; want output only in payloads", output, err)
	}
}

func TestCompleteSchemaViolationKeepsTheSchemaPath(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	s, cfg := openJobs(t), jobConfig()
	raw := strings.Replace(document, `"input":{`, `"max_attempts":1,"input":{"schema":{"type":"object","additionalProperties":{"type":"string"}},`, 1)
	id := submit(t, s, cfg, raw)
	lease := leaseAt(t, &policy.LeaseRequest{Node: "local"}, jobTestStore{s, cfg}, 1)
	report := &jobs.CompletionReport{Outcome: "succeeded", Output: map[string]any{"json": map[string]any{"private-key": 5}}}
	state, err := jobs.CompleteAttempt(ctx, s, cfg, lease.AttemptID, lease.Generation, "local", report, 2)
	if err != nil || state != "failed" {
		t.Fatalf("CompleteAttempt() = %q, %v; want failed", state, err)
	}
	var detail string
	err = s.QueryRowContext(ctx, "SELECT detail FROM results WHERE job_id = ?", id).Scan(&detail)
	if err != nil || detail != `{"error":"schema_violation","schema_path":"/additionalProperties/type"}` {
		t.Fatalf("control detail = %q, %v; want the schema rule", detail, err)
	}
}

func TestTaskQuotaWallRetainsTheProfileAndAdvertisesItsReset(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	data := jobTestStore{openJobs(t), jobConfig()}
	data.cfg.Privacy["mail"] = []string{"ollama", "runner"}
	data.cfg.Profiles = map[string]config.TaskProfile{"task": {Runner: "test-runner", Model: "remote-model", Privacy: []string{"mail"}}}
	queue := data.cfg.Queues["work"]
	queue.Profiles = []string{"task"}
	data.cfg.Queues["work"] = queue
	const raw = `{"contract":1,"kind":"task","privacy":"mail","queue":"work","idempotency_key":"task","input":{"profile":"task","prompt":"test"}}`
	id := submit(t, data.store, data.cfg, raw)
	req := &policy.LeaseRequest{Node: "local", Kind: "task"}
	lease := leaseAt(t, req, data, 1)
	report := &jobs.CompletionReport{Outcome: "failed", ErrorCode: "quota_wall", Usage: map[string]any{"wall_resets_in_s": 7200}}
	state, err := jobs.CompleteAttempt(ctx, data.store, data.cfg, lease.AttemptID, lease.Generation, "local", report, 2)
	if err != nil || state != "queued" {
		t.Fatalf("CompleteAttempt() = %q, %v; want queued", state, err)
	}
	requireNoLease(t, data, req, 7201)
	job, err := jobs.GetJob(ctx, data.store, data.cfg, "producer", id, 3)
	if err != nil || job.CoolingUntil == nil || *job.CoolingUntil != 7202 {
		t.Fatalf("GetJob() = %v, %v; want cooling until 7202", job, err)
	}
	if got := leaseAt(t, req, data, 7202); got.Model != "task" {
		t.Fatalf("LeaseJob().Model = %q; want task profile", got.Model)
	}
}

func TestCompletionOnlyStoresExecutorMetadata(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	data := jobTestStore{openJobs(t), jobConfig()}
	id := submit(t, data.store, data.cfg, document)
	lease := leaseAt(t, localLeaseRequest(), data, 1)
	report := &jobs.CompletionReport{
		Outcome: "succeeded", Output: map[string]any{"text": "answer"},
		Executor: map[string]any{"node": "private node text", "provider": "ollama", "model": "private model text"},
	}
	complete(t, data, lease.AttemptID, lease.Generation, report, 2)
	var executor string
	if err := data.store.QueryRowContext(ctx, "SELECT executor FROM results WHERE job_id = ?", id).Scan(&executor); err != nil {
		t.Fatal(err)
	}
	if executor != `{"node":"local","provider":"ollama"}` {
		t.Fatalf("executor = %q; want the fenced node and known provider", executor)
	}
}
