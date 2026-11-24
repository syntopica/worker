package jobs

import (
	"strings"
	"testing"

	"github.com/syntopica/worker/internal/store"
)

func TestAdminJobQueryUsesIndexedSeeks(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	dir := t.TempDir()
	if err := store.Migrate(ctx, dir); err != nil {
		t.Fatal(err)
	}
	s, err := store.Open(ctx, dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := s.Close(); err != nil {
			t.Error(err)
		}
	})
	cases := map[string]params{
		"detail": {keyID: "unknown", keyLimit: 1},
		"cursor": {"before_time": 0, "before_id": "unknown", keyLimit: 10},
	}
	for name, values := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			steps, err := queryMapRows(t.Context(), s, "EXPLAIN QUERY PLAN "+adminJobQuery(values), values)
			if err != nil {
				t.Fatal(err)
			}
			for _, step := range steps {
				if detail, _ := step["detail"].(string); strings.HasPrefix(detail, "SEARCH j ") {
					return
				}
			}
			t.Fatalf("admin query plan = %v; want a seek on jobs", steps)
		})
	}
}
