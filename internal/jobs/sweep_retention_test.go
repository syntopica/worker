package jobs_test

import (
	"testing"

	"github.com/syntopica/worker/internal/jobs"
)

func TestRetentionExpiresAnUnacknowledgedSensitiveSuccess(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	data := jobTestStore{openJobs(t), jobConfig()}
	queue := data.cfg.Queues["work"]
	queue.RetentionDays, queue.UnackedTTLHours = 7, 1
	data.cfg.Queues["work"] = queue
	id := submit(t, data.store, data.cfg, document)
	lease := leaseAt(t, localLeaseRequest(), data, 1)
	report := &jobs.CompletionReport{Outcome: "succeeded", Output: map[string]any{"text": "answer"}}
	if _, err := jobs.CompleteAttempt(ctx, data.store, data.cfg, lease.AttemptID, lease.Generation, "local", report, 2); err != nil {
		t.Fatal(err)
	}
	if count, err := jobs.SweepRetention(ctx, data.store, data.cfg, 3603); err != nil || count != 1 {
		t.Fatalf("SweepRetention() = %d, %v; want 1", count, err)
	}
	job, err := jobs.GetJob(ctx, data.store, data.cfg, "producer", id, 3603)
	if err != nil || job.State != "unacked_expired" {
		t.Fatalf("GetJob() = %v, %v; want unacked_expired", job, err)
	}
	assertExpiredResult(t, job.Result)
	if count, sweepErr := jobs.SweepRetention(ctx, data.store, data.cfg, 3604); sweepErr != nil || count != 0 {
		t.Fatalf("SweepRetention(repeat) = %d, %v; want 0", count, sweepErr)
	}
}

func assertExpiredResult(t *testing.T, result *jobs.JobResult) {
	t.Helper()
	if result == nil || result.Control == nil || *result.Control != "unacked_expired" || result.Output != nil {
		t.Fatalf("result = %v; want an expiration control without output", result)
	}
}
