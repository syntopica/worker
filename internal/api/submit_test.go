package api_test

import (
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/syntopica/worker/internal/contract"
)

func TestProducerAndNodeRoundTrip(t *testing.T) {
	t.Parallel()
	f := openAPI(t)
	job := response[contract.Submitted](t, f.request(t, http.MethodPost, "/v1/jobs", "producer", submission, http.StatusCreated))
	if repeat := response[contract.Submitted](t, f.request(t, http.MethodPost, "/v1/jobs", "producer", submission, http.StatusOK)); repeat.ID != job.ID || repeat.Created {
		t.Fatalf("repeat = %+v, want existing job %s", repeat, job.ID)
	}

	none := f.request(t, http.MethodPost, "/v1/leases", "guest", `{"node":"guest","user_active":false,"current_idle_s":900}`, http.StatusNoContent)
	if none.Body.Len() != 0 {
		t.Fatalf("204 carried a body: %s", none.Body)
	}
	lease := response[contract.Lease](t, f.request(t, http.MethodPost, "/v1/leases", "local",
		`{"node":"local","resident_model":null,"user_active":false,"free_gb":10,"current_idle_s":900}`, http.StatusOK))
	if lease.JobID != job.ID || lease.Generation != 1 || !strings.Contains(string(lease.Input), "fixture input") {
		t.Fatalf("lease = %+v, want generation one of %s", lease, job.ID)
	}
	f.finish(t, &lease)
	f.collect(t, job.ID)
}

func (f fixture) finish(t *testing.T, lease *contract.Lease) {
	t.Helper()
	attempt := "/v1/attempts/" + lease.AttemptID
	generation := strconv.FormatInt(lease.Generation, 10)
	f.request(t, http.MethodPost, attempt+"/heartbeat", "guest", `{"generation":`+generation+`}`, http.StatusConflict)
	f.request(t, http.MethodPost, attempt+"/heartbeat", "local", `{"generation":`+generation+`,"draining":false}`, http.StatusOK)
	outcome := `{"generation":` + generation + `,"outcome":"succeeded","output":{"text":"fixture output","json":null},"usage":{},"executor":{"node":"local"},"error_code":null,"wall_s":1.0}`
	wantJSON(t, f.request(t, http.MethodPost, attempt+"/complete", "local", outcome, http.StatusOK), `"state":"succeeded"`)
	f.request(t, http.MethodPost, attempt+"/complete", "local", outcome, http.StatusConflict)
}

func (f fixture) collect(t *testing.T, id string) {
	t.Helper()
	path := "/v1/jobs/" + id
	f.request(t, http.MethodGet, path, "other", "", http.StatusNotFound)
	wantJSON(t, f.request(t, http.MethodGet, path, "producer", "", http.StatusOK), `"cooling_until":null`)
	results := response[struct {
		Results []contract.Result `json:"results"`
	}](t, f.request(t, http.MethodGet, "/v1/results?queue=work&after=0&wait=1", "producer", "", http.StatusOK))
	if len(results.Results) != 1 || !strings.Contains(string(results.Results[0].Output), "fixture output") {
		t.Fatalf("results = %+v, want the completed output", results)
	}
	f.request(t, http.MethodPost, path+"/ack", "producer", `{"result_id":"`+results.Results[0].ResultID+`","rating":"good"}`, http.StatusOK)
	wantJSON(t, f.request(t, http.MethodGet, "/v1/results?queue=work", "producer", "", http.StatusOK), `"results":[]`)
}

func TestProducerCanCancelOwnedJob(t *testing.T) {
	t.Parallel()
	f := openAPI(t)
	job := response[contract.Submitted](t, f.request(t, http.MethodPost, "/v1/jobs", "producer", submission, http.StatusCreated))
	path := "/v1/jobs/" + job.ID + "/cancel"
	f.request(t, http.MethodPost, path, "other", "", http.StatusNotFound)
	wantJSON(t, f.request(t, http.MethodPost, path, "producer", "", http.StatusOK), `"state":"cancelled"`)
}
