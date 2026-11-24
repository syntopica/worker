package jobs_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/syntopica/worker/internal/jobs"
)

func TestRetryAcknowledgesTheOriginalWithinItsOutstandingLimit(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	data := jobTestStore{openJobs(t), jobConfig()}
	queue := data.cfg.Queues["work"]
	queue.MaxOutstanding = 1
	data.cfg.Queues["work"] = queue
	id := submit(t, data.store, data.cfg, strings.Replace(document, `"kind"`, `"max_attempts":1,"kind"`, 1))
	lease := leaseAt(t, localLeaseRequest(), data, 1)
	report := &jobs.CompletionReport{Outcome: "failed", ErrorCode: "unrecognized private text"}
	state, err := jobs.CompleteAttempt(ctx, data.store, data.cfg, lease.AttemptID, lease.Generation, "local", report, 2)
	if err != nil || state != "failed" {
		t.Fatalf("CompleteAttempt() = %q, %v; want failed", state, err)
	}
	job, err := jobs.GetJob(ctx, data.store, data.cfg, "producer", id, 2)
	if err != nil || job.Error == nil || *job.Error != "executor_error" {
		t.Fatalf("GetJob() = %v, %v; want the allowlisted code", job, err)
	}
	retry, err := jobs.RetryJob(ctx, data.store, data.cfg, id, "operator", 3)
	if err != nil || !retry.Created || retry.ID == id {
		t.Fatalf("RetryJob() = %v, %v; want a new job", retry, err)
	}
	_, err = jobs.ReadAdminContent(ctx, data.store, id, "mail", "operator", 4)
	requireRefusal(t, err, http.StatusGone, "content_gone")
	if got := leaseAt(t, localLeaseRequest(), data, 4); got.JobID != retry.ID {
		t.Fatalf("LeaseJob().JobID = %q, want retry %q", got.JobID, retry.ID)
	}
}
