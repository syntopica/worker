package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/syntopica/worker/internal/contract"
)

func TestAdminSensitiveReadRequiresRevealAndAudit(t *testing.T) {
	t.Parallel()
	f := openAPI(t)
	job := response[contract.Submitted](t, f.request(t, http.MethodPost, "/v1/jobs", "producer", submission, http.StatusCreated))
	path := "/v1/admin/jobs/" + job.ID
	wantJSON(t, f.request(t, http.MethodGet, path, "admin", "", http.StatusOK), `"has_input":true`)
	f.request(t, http.MethodGet, path+"/content", "producer", "", http.StatusForbidden)
	wantJSON(t, f.request(t, http.MethodGet, path+"/content", "admin", "", http.StatusForbidden), "reveal_required")
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, path+"/content", nil)
	req.Header.Set("Authorization", "Bearer "+f.tokens["admin"])
	req.Header.Set("X-Worker-Reveal", "mail")
	wantJSON(t, f.send(t, req, http.StatusOK), "fixture input")
	wantJSON(t, f.request(t, http.MethodPost, path+"/cancel", "admin", "", http.StatusOK), "cancelled")
	wantJSON(t, f.send(t, req, http.StatusGone), "content_gone")
	audit := f.request(t, http.MethodGet, "/v1/admin/audit?days=0", "admin", "", http.StatusOK)
	wantJSON(t, audit, `"action":"reveal"`)
	wantJSON(t, audit, `"action":"cancel"`)
}

func TestAdminReportsAndFilters(t *testing.T) {
	t.Parallel()
	f := openAPI(t)
	for path, want := range map[string]string{
		"/v1/status":             `"openrouter":{`,
		"/v1/activity?hours=168": `"bucket_s":21600`,
		"/v1/costs?days=5":       `"rows":[]`,
		"/v1/quality?days=5":     `"judged":[]`,
		"/v1/admin/jobs?queue=work&state=queued,new":  "unknown_state",
		"/v1/admin/jobs?queue=no":                     "unknown_queue",
		"/v1/admin/jobs?state=leased,running&limit=0": `"jobs":[]`,
	} {
		status := http.StatusOK
		if want == "unknown_state" || want == "unknown_queue" {
			status = http.StatusBadRequest
		}
		f.request(t, http.MethodGet, path, "producer", "", http.StatusForbidden)
		wantJSON(t, f.request(t, http.MethodGet, path, "admin", "", status), want)
	}
}
