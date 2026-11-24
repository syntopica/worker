package jobs_test

import (
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/syntopica/worker/internal/config"
	"github.com/syntopica/worker/internal/contract"
	"github.com/syntopica/worker/internal/fault"
	"github.com/syntopica/worker/internal/jobs"
	"github.com/syntopica/worker/internal/policy"
	"github.com/syntopica/worker/internal/store"
)

const document = `{"contract":1,"kind":"inference","privacy":"mail","queue":"work","idempotency_key":"x","requirements":{"models":["model"]},"input":{"messages":[{"role":"user","content":"test"}]}}`

func openJobs(t *testing.T) *store.Store {
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
	return s
}

func jobConfig() *config.Config {
	return &config.Config{
		MaxPayloadBytes: 10000,
		Producers:       map[string][]string{"producer": {"work"}},
		Models:          map[string]config.ModelPin{"model": {ColdGB: 2}},
		Nodes: map[string]config.NodePolicy{
			"local": {Trust: "owner", MemoryBudgetGB: 10},
			"other": {Trust: "guest", MemoryBudgetGB: 10},
		},
		Queues: map[string]config.QueuePolicy{
			"work": {RunWhen: "idle", Weight: 1, MaxOutstanding: 10},
		},
		Privacy: map[string][]string{"mail": {"ollama"}},
		Trust:   map[string][]string{"mail": {"owner"}},
	}
}

func submit(t *testing.T, s *store.Store, cfg *config.Config, raw string) string {
	t.Helper()
	id, created, err := jobs.SubmitJob(t.Context(), s, cfg, "producer", []byte(raw), 0)
	if err != nil || !created || id == "" {
		t.Fatalf("SubmitJob() = %q, %t, %v; want a new job", id, created, err)
	}
	return id
}

func requireRefusal(t *testing.T, err error, status int, code string) {
	t.Helper()
	var refusal *fault.Error
	if !errors.As(err, &refusal) || refusal.Status != status || refusal.Code != code {
		t.Fatalf("got error %v, want %d %s", err, status, code)
	}
}

func TestLeaseAndFence(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	data := jobTestStore{openJobs(t), jobConfig()}
	id := submit(t, data.store, data.cfg, document)
	req := &policy.LeaseRequest{Node: "other", CurrentIdle: 900}
	requireNoLease(t, data, req, 1)
	req.Node = "local"
	first := leaseAt(t, req, data, 1)
	if first.JobID != id || first.Generation != 1 || first.Input == nil {
		t.Fatalf("LeaseJob() = %v; want generation 1 for %s", first, id)
	}
	requireNoLease(t, data, req, 2)
	requireRefusal(t, jobs.HeartbeatAttempt(ctx, data.store, first.AttemptID, first.Generation, "other", false, 2),
		http.StatusConflict, "stale_attempt")
	if count, err := jobs.ExpireLeases(ctx, data.store, 62); err != nil || count != 1 {
		t.Fatalf("ExpireLeases() = %d, %v; want 1", count, err)
	}
	second := leaseAt(t, req, data, 70)
	if second.Generation != 2 || second.AttemptID == first.AttemptID {
		t.Fatalf("LeaseJob() after expiry = %v; want a new attempt with generation 2", second)
	}
	requireRefusal(t, jobs.HeartbeatAttempt(ctx, data.store, first.AttemptID, first.Generation, "local", false, 71),
		http.StatusConflict, "stale_attempt")
	assertHeartbeat(t, data, second, 71)
}

func assertHeartbeat(t *testing.T, data jobTestStore, lease *contract.Lease, now float64) {
	t.Helper()
	ctx := t.Context()
	if err := jobs.HeartbeatAttempt(ctx, data.store, lease.AttemptID, lease.Generation, "local", true, now); err != nil {
		t.Fatal(err)
	}
	var state string
	var expires float64
	if err := data.store.QueryRowContext(ctx, "SELECT state, lease_expires FROM jobs WHERE id = ?", lease.JobID).Scan(&state, &expires); err != nil {
		t.Fatal(err)
	}
	if state != "draining" || expires != now+60 {
		t.Fatalf("heartbeat stored %q, %v; want draining, %v", state, expires, now+60)
	}
}

func requireNoLease(t *testing.T, data jobTestStore, request *policy.LeaseRequest, now float64) {
	t.Helper()
	if lease, err := jobs.LeaseJob(t.Context(), data.store, data.cfg, request, now); err != nil || lease != nil {
		t.Fatalf("LeaseJob() = %v, %v; want no lease", lease, err)
	}
}

type jobTestStore struct {
	store *store.Store
	cfg   *config.Config
}

func localLeaseRequest() *policy.LeaseRequest {
	return &policy.LeaseRequest{Node: "local", CurrentIdle: 900}
}

func TestLeaseSkipsAJobWithLostPayload(t *testing.T) {
	t.Parallel()
	data := jobTestStore{openJobs(t), jobConfig()}
	ctx := t.Context()
	missing := submit(t, data.store, data.cfg, document)
	next := submit(t, data.store, data.cfg, strings.Replace(document, `"x"`, `"next"`, 1))
	if _, err := data.store.ExecContext(ctx, "DELETE FROM p.inputs WHERE job_id = ?", missing); err != nil {
		t.Fatal(err)
	}
	lease := leaseAt(t, localLeaseRequest(), data, 1)
	if lease.JobID != next {
		t.Fatalf("LeaseJob().JobID = %q; want %q", lease.JobID, next)
	}
	var detail string
	if err := data.store.QueryRowContext(ctx, "SELECT detail FROM results WHERE job_id = ?", missing).Scan(&detail); err != nil {
		t.Fatal(err)
	}
	if detail != `{"error":"payload_lost"}` {
		t.Fatalf("control detail = %q; want payload_lost", detail)
	}
}

func TestOnDemandLeaseAuditsOverrideAndCountsItsRunner(t *testing.T) {
	t.Parallel()
	data := jobTestStore{openJobs(t), jobConfig()}
	ctx := t.Context()
	data.cfg.Privacy["mail"] = []string{"ollama", "runner"}
	data.cfg.Profiles = map[string]config.TaskProfile{
		"manual": {Runner: "runner-cli", Privacy: []string{"mail"}, OnDemand: true},
	}
	data.cfg.RunnerMaxConcurrent = map[string]int{"runner-cli": 1}
	id := submit(t, data.store, data.cfg, document)
	submit(t, data.store, data.cfg, strings.Replace(document, `"x"`, `"next"`, 1))
	if _, err := data.store.ExecContext(ctx, "UPDATE jobs SET pin = 'ollama:model', priority = 100 WHERE id = ?", id); err != nil {
		t.Fatal(err)
	}
	req := &policy.LeaseRequest{Node: "local", Kind: "task", Profile: "manual", OverridePins: true}
	lease := leaseAt(t, req, data, 1)
	if lease.JobID != id || !lease.Pinned || lease.Model != "manual" || string(lease.Input) != `{"prompt":"[user]\ntest"}` {
		t.Fatalf("LeaseJob() = %v; want pinned job on manual profile", lease)
	}
	requireNoLease(t, data, req, 2)
	assertOverrideAudit(t, data, id)
}

func assertOverrideAudit(t *testing.T, data jobTestStore, id string) {
	t.Helper()
	audit, err := jobs.ReadAudit(t.Context(), data.store, 0)
	if err != nil || len(audit) != 1 {
		t.Fatalf("ReadAudit() = %v, %v; want one action", audit, err)
	}
	if audit[0]["action"] != "override_pin" || audit[0]["job_id"] != id {
		t.Fatalf("audit = %v; want one override_pin action for %q", audit, id)
	}
}

func TestLeaseRefusesUnknownNode(t *testing.T) {
	t.Parallel()
	req := &policy.LeaseRequest{Node: "unknown"}
	_, err := jobs.LeaseJob(t.Context(), openJobs(t), jobConfig(), req, 1)
	requireRefusal(t, err, http.StatusForbidden, "unknown_node")
}

func TestOpenRouterDailyCapReleasesTheLocalRung(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	data := jobTestStore{openJobs(t), jobConfig()}
	limit := 1
	queue := data.cfg.Queues["work"]
	queue.OpenRouter = &config.OpenRouterRoute{Models: map[string]string{"model": "test/free:free"}, DailyCap: &limit}
	queue.LocalAfter = time.Hour
	data.cfg.Queues["work"] = queue
	data.cfg.Privacy["mail"] = []string{"ollama", "openrouter"}
	id := submit(t, data.store, data.cfg, document)
	remote := &policy.LeaseRequest{Node: "local", Kind: "openrouter"}
	requireNoLease(t, data, localLeaseRequest(), 1)
	lease := leaseAt(t, remote, data, 2)
	if lease.JobID != id || lease.Model != "test/free:free" {
		t.Fatalf("LeaseJob() = %v; want the remote route", lease)
	}
	complete(t, data, lease.AttemptID, lease.Generation, &jobs.CompletionReport{
		Outcome: "succeeded", Output: map[string]any{"text": "answer"},
		Executor: map[string]any{"provider": "openrouter", "model": lease.Model},
	}, 3)
	next := submit(t, data.store, data.cfg, strings.Replace(document, `"x"`, `"next"`, 1))
	requireNoLease(t, data, remote, 4)
	if got := leaseAt(t, localLeaseRequest(), data, 4); got.JobID != next || got.Model != "model" {
		t.Fatalf("LeaseJob() = %v; want the local rung", got)
	}
	headroom, err := jobs.OpenRouterHeadroom(ctx, data.store, data.cfg, 4)
	assertReport(t, headroom, err, `"capped":["work"]`)
}
