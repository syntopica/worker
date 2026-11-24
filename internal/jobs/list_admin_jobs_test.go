package jobs_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/syntopica/worker/internal/jobs"
)

func TestAdminCursorPaginatesTiedCreationTimes(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	data := jobTestStore{openJobs(t), jobConfig()}
	first := submit(t, data.store, data.cfg, document)
	second := submit(t, data.store, data.cfg, strings.Replace(document, `"x"`, `"second"`, 1))
	want := jobs.Filter{Limit: 1, States: []string{"queued"}}
	page, err := jobs.ListAdminJobs(ctx, data.store, want)
	if err != nil {
		t.Fatal(err)
	}
	cursor, ok := page["next"].(string)
	if !ok || !strings.HasPrefix(cursor, "0.0:") {
		t.Fatalf("next cursor = %v; want the Python cursor format", page["next"])
	}
	want.Before = &cursor
	next, err := jobs.ListAdminJobs(ctx, data.store, want)
	if err != nil {
		t.Fatal(err)
	}
	pageIDs := asJSON(t, page["jobs"]) + asJSON(t, next["jobs"])
	if !strings.Contains(pageIDs, first) || !strings.Contains(pageIDs, second) {
		t.Fatalf("pages did not contain both jobs: %s", pageIDs)
	}
	invalid := "abc"
	want.Before = &invalid
	_, err = jobs.ListAdminJobs(ctx, data.store, want)
	requireRefusal(t, err, http.StatusBadRequest, "bad_cursor")
}
