package jobs_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/syntopica/worker/internal/jobs"
)

func TestSubmitValidation(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		raw  string
		code string
	}{
		"non-object":         {"[]", "bad_json"},
		"trailing JSON":      {document + " {}", "bad_json"},
		"contract":           {strings.Replace(document, "1", "2", 1), "unsupported_contract"},
		"kind":               {strings.Replace(document, "inference", "unsupported", 1), "unsupported_kind"},
		"privacy":            {strings.Replace(document, "mail", "unknown", 1), "unknown_privacy"},
		"tier":               {strings.Replace(document, `"kind"`, `"tier":"unknown","kind"`, 1), "unknown_tier"},
		"empty key":          {strings.Replace(document, `"x"`, `""`, 1), "bad_idempotency_key"},
		"boolean priority":   {strings.Replace(document, `"kind"`, `"priority":true,"kind"`, 1), "out_of_range"},
		"large priority":     {strings.Replace(document, `"kind"`, `"priority":101,"kind"`, 1), "out_of_range"},
		"empty split parent": {strings.Replace(document, `"kind"`, `"parent_id":"","kind"`, 1), "bad_split_child"},
		"invalid messages":   {strings.Replace(document, `"content":"test"`, `"content":7`, 1), "bad_input"},
		"invalid exclusion":  {strings.Replace(document, `"models"`, `"exclude_providers":[3],"models"`, 1), "bad_requirements"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, _, err := jobs.SubmitJob(t.Context(), openJobs(t), jobConfig(), "producer", []byte(tc.raw), 0)
			requireRefusal(t, err, http.StatusBadRequest, tc.code)
		})
	}
}

func TestSubmitReusesThePythonHash(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	s, cfg := openJobs(t), jobConfig()
	raw := strings.Replace(document, "test", `á🦫\u007f\n`, 1)
	raw = strings.Replace(raw, `"kind"`, `"options":{"a":1e-5,"b":1e16,"c":-0.0,"d":1.0},"kind"`, 1)
	id := submit(t, s, cfg, raw)
	var hash string
	if err := s.QueryRowContext(ctx, "SELECT payload_hash FROM jobs WHERE id = ?", id).Scan(&hash); err != nil {
		t.Fatal(err)
	}
	// Produced by Python payload_hash on the same decoded submission.
	const pythonHash = "3b84f8800302af228fb1a3fea9bbf8cc676559ac24cda8170de68963e17f7bfa"
	if hash != pythonHash {
		t.Fatalf("payload hash = %q; want %q", hash, pythonHash)
	}
	repeat, created, err := jobs.SubmitJob(ctx, s, cfg, "producer", []byte(raw), 1)
	if err != nil || created || repeat != id {
		t.Fatalf("SubmitJob(repeat) = %q, %t, %v; want %q, false", repeat, created, err, id)
	}
	_, _, err = jobs.SubmitJob(ctx, s, cfg, "producer", []byte(document), 2)
	requireRefusal(t, err, http.StatusConflict, "idempotency_conflict")
}

func TestSubmitExcludingEveryProviderFailsClosed(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	data := jobTestStore{openJobs(t), jobConfig()}
	raw := strings.Replace(document, `"models"`, `"exclude_providers":["ollama","ollama"],"models"`, 1)
	id := submit(t, data.store, data.cfg, raw)
	requireNoLease(t, data, localLeaseRequest(), 1)
	var detail string
	if err := data.store.QueryRowContext(ctx, "SELECT detail FROM results WHERE job_id = ?", id).Scan(&detail); err != nil {
		t.Fatal(err)
	}
	if detail != `{"error":"no_eligible_executor"}` {
		t.Fatalf("control detail = %q; want no_eligible_executor", detail)
	}
}
