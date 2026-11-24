package jobs_test

import (
	"strings"
	"testing"

	"github.com/syntopica/worker/internal/jobs"
)

func TestReportsKeepUsageAndOmitPayload(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	data := jobTestStore{openJobs(t), jobConfig()}
	submit(t, data.store, data.cfg, document)
	lease := leaseAt(t, localLeaseRequest(), data, 1)
	complete(t, data, lease.AttemptID, lease.Generation, &jobs.CompletionReport{
		Outcome: "succeeded", Output: map[string]any{"text": "only in payload store"},
		Usage:    map[string]any{"tokens_in": 4, "cost_usd": 0.01},
		Executor: map[string]any{"provider": "ollama", "model": "model"}, WallS: 1,
	}, 2)
	report, err := jobs.ReadStatus(ctx, data.store, 3)
	assertReport(t, report, err, `"done_1h":1`)
	report, err = jobs.ReadActivity(ctx, data.store, 3, 1)
	assertReport(t, report, err, `"sampling":false`)
	costs, err := jobs.ReadCosts(ctx, data.store, 0)
	assertReport(t, costs, err, `"cost_usd":0.01`)
	quality, err := jobs.ReadAttemptQuality(ctx, data.store, 0)
	assertReport(t, quality, err, `"succeeded":1`)
	ratings, err := jobs.ReadResultRatings(ctx, data.store, 0)
	assertReport(t, ratings, err, `"results":1`)
}

func assertReport(t *testing.T, value any, err error, expected string) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	raw := asJSON(t, value)
	if !strings.Contains(raw, expected) || strings.Contains(raw, "only in payload store") {
		t.Fatalf("report = %s; want %s and no output content", raw, expected)
	}
}
