package client_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/syntopica/worker/internal/client"
)

func TestAdminReportDoesNotConvertJSONNumbers(t *testing.T) {
	t.Parallel()
	const body = `{"attempts":[{"tokens_in":9007199254740993}],"ratings":[],"judged":[]}`
	server := httptest.NewServer(http.HandlerFunc(func(out http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet || request.URL.Path != "/v1/quality" || request.URL.Query().Get("days") != "0.25" {
			t.Error("unexpected report request")
		}
		if request.Header.Get("Authorization") != "Bearer fixture" {
			t.Error("report authentication missing")
		}
		if _, err := io.WriteString(out, body); err != nil {
			t.Error(err)
		}
	}))
	defer server.Close()

	query := url.Values{"days": {"0.25"}}
	admin := client.New(server.URL, "fixture", server.Client())
	raw, err := admin.AdminReport(t.Context(), "/v1/quality?"+query.Encode())
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != body {
		t.Fatal("admin report changed the response")
	}
}
