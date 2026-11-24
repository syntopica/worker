package api_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/syntopica/worker/internal/config"
)

func TestRequestRefusals(t *testing.T) {
	t.Parallel()
	f := openAPI(t)
	cases := []struct {
		method, path, principal, body, code string
		status                              int
	}{
		{http.MethodGet, "/v1/status", "unknown", "", "unauthorized", 401},
		{http.MethodGet, "/v1/no-route", "admin", "", "no_route", 404},
		{http.MethodPost, "/v1/leases", "producer", "{bad", "forbidden", 403},
		{http.MethodPost, "/v1/leases", "local", `{"node":"guest"}`, "forbidden", 403},
		{http.MethodPost, "/v1/leases", "local", `{"node":"local","kind":"other"}`, "unsupported_kind", 400},
		{http.MethodPost, "/v1/leases", "local", `{"node":"local","kind":"task","profile":"other"}`, "unknown_profile", 400},
		{http.MethodPost, "/v1/leases", "local", `{"node":"local","user_active":false,"current_idle_s":900,"created_before":1}`, "bad_request", 400},
		{http.MethodPost, "/v1/leases", "local", `{"node":"local","user_active":false,"current_idle_s":900,"override_pins":"true"}`, "bad_request", 400},
		{http.MethodPost, "/v1/jobs", "producer", "[0]", "bad_json", 400},
		{http.MethodPost, "/v1/jobs/id/ack", "producer", "{} {}", "bad_request", 400},
		{http.MethodPost, "/v1/jobs/id/ack", "producer", `{"rating":"unknown"}`, "unknown_rating", 400},
		{http.MethodPost, "/v1/attempts/id/heartbeat", "local", `{"generation":null}`, "bad_request", 400},
		{http.MethodPost, "/v1/attempts/id/complete", "local", `{"generation":1}`, "bad_request", 400},
		{http.MethodGet, "/v1/results?queue=work&wait=nan", "producer", "", "bad_request", 400},
		{http.MethodGet, "/v1/results?queue=work", "other", "", "queue_not_granted", 403},
		{http.MethodGet, "/v1/activity?hours=no", "admin", "", "bad_request", 400},
		{http.MethodGet, "/v1/costs?days=nan", "admin", "", "bad_request", 400},
	}
	for _, tc := range cases {
		wantJSON(t, f.request(t, tc.method, tc.path, tc.principal, tc.body, tc.status), `"error": "`+tc.code+`"`)
	}
}

func TestBodyLimitAppliesToDeclaredAndStreamedBodies(t *testing.T) {
	t.Parallel()
	f := openAPI(t)
	chunked := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/jobs", strings.NewReader(strings.Repeat("x", f.cfg.MaxPayloadBytes+1)))
	chunked.ContentLength = -1
	wantJSON(t, f.send(t, chunked, http.StatusRequestEntityTooLarge), "payload_too_large")
	declared := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/v1/status", nil)
	declared.ContentLength = int64(f.cfg.MaxPayloadBytes + 1)
	wantJSON(t, f.send(t, declared, http.StatusRequestEntityTooLarge), "payload_too_large")
	f.request(t, http.MethodPost, "/v1/jobs", "producer", strings.Repeat("x", f.cfg.MaxPayloadBytes), http.StatusBadRequest)
}

func TestNodeReportsAndCooldowns(t *testing.T) {
	t.Parallel()
	f := openAPI(t)
	f.cfg.Profiles = map[string]config.TaskProfile{"fixture": {Runner: "codex", Model: "fixture-model"}}
	f.cfg.RunnerCooldown = map[string]time.Duration{"agy": time.Hour}
	path := "/v1/nodes/local/"
	f.request(t, http.MethodPost, path+"report", "guest", `{"reason":"idle"}`, http.StatusForbidden)
	f.request(t, http.MethodPost, path+"report", "local", "null", http.StatusBadRequest)
	f.request(t, http.MethodPost, path+"report", "local", "{\"reason\":\"\xff\"}", http.StatusBadRequest)
	for _, raw := range []string{
		`{"extra":0}`, `{"PRESSURE":"normal"}`, `{"pressure":"other"}`, `{"pressure":null}`, `{"reason":"other"}`,
		`{"on_ac":1}`, `{"idle_s":-1}`, `{"loads_1h":-1}`, `{"loads_1h":null}`, `{"unloads_1h":1.5}`,
		`{"resident":[1]}`, `{"resident":[null]}`, `{"unexpected":null}`,
	} {
		f.request(t, http.MethodPost, path+"report", "local", raw, http.StatusBadRequest)
	}
	f.request(t, http.MethodPost, path+"report", "local", `{}`, http.StatusOK)
	partial := f.request(t, http.MethodGet, "/v1/status", "admin", "", http.StatusOK)
	wantJSON(t, partial, `"pressure":"unknown"`)
	wantJSON(t, partial, `"resident":[]`)
	f.request(t, http.MethodPost, path+"report", "local",
		`{"reason":"on_battery","idle_s":12.5,"on_ac":false,"pressure":"normal","resident":["model","other-model"],"unexpected":["other-model"],"loads_1h":2,"unloads_1h":1}`, http.StatusOK)
	f.request(t, http.MethodPost, path+"walls", "guest", "{}", http.StatusForbidden)
	f.request(t, http.MethodPost, path+"walls", "local", `{"codex":"tomorrow"}`, http.StatusBadRequest)
	f.request(t, http.MethodPost, path+"walls", "local", `{"codex":9999999999}`, http.StatusBadRequest)
	f.request(t, http.MethodPost, path+"walls", "local", `{"codex":1,"codex:fixture-model":1}`, http.StatusOK)
	reset := strconv.FormatInt(time.Now().Add(time.Hour).Unix(), 10)
	f.request(t, http.MethodPost, path+"walls", "local", `{"codex":`+reset+`,"unknown":`+reset+`}`, http.StatusBadRequest)
	f.request(t, http.MethodPost, path+"walls", "local", `{"codex:other-model":`+reset+`}`, http.StatusBadRequest)
	status := f.request(t, http.MethodGet, "/v1/status", "admin", "", http.StatusOK)
	wantJSON(t, status, `"reason":"on_battery"`)
	wantJSON(t, status, `"resident":["model","unconfigured"]`)
	wantJSON(t, status, `"unexpected":["unconfigured"]`)
	wantJSON(t, status, `"cooldowns":{}`)
	f.request(t, http.MethodPost, path+"walls", "local", `{"agy":`+reset+`,"codex:fixture-model":`+reset+`}`, http.StatusOK)
	status = f.request(t, http.MethodGet, "/v1/status", "admin", "", http.StatusOK)
	wantJSON(t, status, `"cooldowns":{"agy":`)
	wantJSON(t, status, `"codex:fixture-model":`)
}
