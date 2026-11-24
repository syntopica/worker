package jobs_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/syntopica/worker/internal/jobs"
)

func TestAdminMetadataAndSensitiveReveal(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	data := jobTestStore{openJobs(t), jobConfig()}
	raw := strings.Replace(document, "test", "private prompt", 1)
	id := submit(t, data.store, data.cfg, raw)
	lease := leaseAt(t, localLeaseRequest(), data, 1)
	report := &jobs.CompletionReport{
		Outcome: "succeeded", Output: map[string]any{"text": "private answer"},
		Usage:    map[string]any{"tokens_in": 2, "unknown": "private diagnostic"},
		Executor: map[string]any{"provider": "ollama", "model": "model", "diagnostic": "private diagnostic"},
	}
	complete(t, data, lease.AttemptID, lease.Generation, report, 2)
	assertAdminMetadata(t, data, id)
	_, err := jobs.ReadAdminContent(ctx, data.store, id, "personal", "operator", 3)
	requireRefusal(t, err, http.StatusForbidden, "reveal_required")
	content, err := jobs.ReadAdminContent(ctx, data.store, id, "mail", "operator", 3)
	if err != nil || !strings.Contains(asJSON(t, content), "private answer") {
		t.Fatalf("ReadAdminContent() = %v, %v; want stored content", content, err)
	}
	audit, err := jobs.ReadAudit(ctx, data.store, 0)
	if err != nil || len(audit) != 1 || audit[0]["action"] != "reveal" {
		t.Fatalf("ReadAudit() = %v, %v; want one reveal", audit, err)
	}
}

func assertAdminMetadata(t *testing.T, data jobTestStore, id string) {
	t.Helper()
	ctx := t.Context()
	detail, err := jobs.ReadAdminJob(ctx, data.store, id)
	if err != nil {
		t.Fatal(err)
	}
	if detail["has_input"] != true || detail["has_output"] != true || detail["tokens_in"] != int64(2) {
		t.Fatalf("ReadAdminJob() = %v; want only payload presence and attempt totals", detail)
	}
	if raw := asJSON(t, detail); strings.Contains(raw, "private") {
		t.Fatalf("ReadAdminJob() included payload content: %s", raw)
	}
}

func asJSON(t *testing.T, value any) string {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}
