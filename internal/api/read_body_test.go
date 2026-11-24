package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
	path := "/v1/nodes/local/"
	f.request(t, http.MethodPost, path+"report", "guest", `{"reason":"idle"}`, http.StatusForbidden)
	f.request(t, http.MethodPost, path+"report", "local", "null", http.StatusBadRequest)
	f.request(t, http.MethodPost, path+"report", "local", "{\"reason\":\"\xff\"}", http.StatusBadRequest)
	f.request(t, http.MethodPost, path+"report", "local", `{"reason":"idle"}`, http.StatusOK)
	f.request(t, http.MethodPost, path+"walls", "guest", "{}", http.StatusForbidden)
	f.request(t, http.MethodPost, path+"walls", "local", `{"runner":"tomorrow"}`, http.StatusBadRequest)
	f.request(t, http.MethodPost, path+"walls", "local", `{"runner":9999999999}`, http.StatusBadRequest)
	f.request(t, http.MethodPost, path+"walls", "local", `{"runner":1}`, http.StatusOK)
	status := f.request(t, http.MethodGet, "/v1/status", "admin", "", http.StatusOK)
	wantJSON(t, status, `"reason":"idle"`)
	wantJSON(t, status, `"cooldowns":{}`)
}
