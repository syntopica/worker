package jobs_test

import (
	"net/http"
	"testing"

	"github.com/syntopica/worker/internal/contract"
	"github.com/syntopica/worker/internal/jobs"
)

func TestProducerCollectsAndAcknowledgesASensitiveResult(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	data := jobTestStore{openJobs(t), jobConfig()}
	id := submit(t, data.store, data.cfg, document)
	lease := leaseAt(t, localLeaseRequest(), data, 1)
	report := &jobs.CompletionReport{
		Outcome: "succeeded", Output: map[string]any{"text": "answer"},
		Usage: map[string]any{"tokens_in": 2, "cost_usd": 0.01},
	}
	if _, err := jobs.CompleteAttempt(ctx, data.store, data.cfg, lease.AttemptID, lease.Generation, "local", report, 2); err != nil {
		t.Fatal(err)
	}
	result := collectedResult(t, data, id)
	ack := contract.Ack{ResultID: result.ResultID, Rating: "edited"}
	requireRefusal(t, jobs.AckResult(ctx, data.store, data.cfg, "another", id, ack, 3), http.StatusNotFound, "not_found")
	if err := jobs.AckResult(ctx, data.store, data.cfg, "producer", id, ack, 3); err != nil {
		t.Fatal(err)
	}
	assertAcked(t, data, id, result.Seq)
}

func collectedResult(t *testing.T, data jobTestStore, id string) jobs.JobResult {
	t.Helper()
	ctx := t.Context()
	results, err := jobs.ListResults(ctx, data.store, "producer", "work", 0, 100)
	if err != nil || len(results) != 1 {
		t.Fatalf("ListResults() = %v, %v; want one result", results, err)
	}
	job, err := jobs.GetJob(ctx, data.store, data.cfg, "producer", id, 2)
	if err != nil || job.Result == nil || results[0].ResultID != job.Result.ResultID {
		t.Fatalf("GetJob() = %v, %v; want the completed result", job, err)
	}
	if string(job.Result.Usage) != `{"cost_usd":0.01,"tokens_in":2}` {
		t.Fatalf("result usage = %s; want the recorded measurements", job.Result.Usage)
	}
	return results[0]
}

func assertAcked(t *testing.T, data jobTestStore, id string, seq int64) {
	t.Helper()
	ctx := t.Context()
	job, err := jobs.GetJob(ctx, data.store, data.cfg, "producer", id, 4)
	if err != nil || job.Result != nil {
		t.Fatalf("GetJob(acked) = %v, %v; want a job with no unacknowledged result", job, err)
	}
	var rating string
	var payloads int
	err = data.store.QueryRowContext(ctx,
		"SELECT rating, (SELECT count(*) FROM p.inputs) + (SELECT count(*) FROM p.outputs) FROM results WHERE seq = ?", seq,
	).Scan(&rating, &payloads)
	if err != nil || rating != "edited" || payloads != 0 {
		t.Fatalf("ack stored %q rating, %d payloads, %v; want edited and no payloads", rating, payloads, err)
	}
}
