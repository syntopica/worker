package store

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"
)

// rawStore builds a store of an older shape directly, the way a previous
// engine version left it.
func rawStore(t *testing.T, dir, schema string, statements ...string) {
	t.Helper()

	db, err := connect(dir)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = db.Close() }()

	err = runStatements(t.Context(), db, schema+";"+strings.Join(statements, ";"))
	if err != nil {
		t.Fatal(err)
	}
}

const legacyJob = "INSERT INTO jobs (id, producer, queue, kind, idempotency_key, payload_hash, priority," +
	" privacy, model, state, max_attempts, not_before, created, updated, finished) VALUES "

func TestTheLegacyStoreGainsTheColumnAndKeepsItsRows(t *testing.T) {
	t.Parallel()

	legacy := strings.Replace(schemaSQL, "  payloads_deleted INTEGER NOT NULL DEFAULT 0,\n", "", 1)
	if strings.Contains(legacy, "payloads_deleted") {
		t.Fatal("the legacy schema still has the column")
	}

	dir := filepath.Join(t.TempDir(), "state")
	rawStore(t, dir, legacy,
		legacyJob+"('old','pa','pa.bulk','inference','k','h',50,'mail','model-a','failed',3,0,0,0,1)",
		legacyJob+"('live','pa','pa.bulk','inference','k2','h',50,'mail','model-a','queued',3,0,0,0,NULL)",
		"INSERT INTO p.inputs (job_id, body) VALUES ('live', '{}')",
	)

	_, err := Open(t.Context(), dir)
	if err == nil {
		t.Fatal("a legacy store opened without a migration")
	}

	for range 2 { // idempotent on the second migration
		opened := freshStore(t, dir)
		if got := scan[int](t, opened, "SELECT payloads_deleted FROM jobs WHERE id='old'"); got != 1 {
			t.Fatalf("old payloads_deleted = %d", got)
		}

		if got := scan[int](t, opened, "SELECT payloads_deleted FROM jobs WHERE id='live'"); got != 0 {
			t.Fatalf("live payloads_deleted = %d", got)
		}

		if got := scan[int](t, opened, "SELECT count(*) FROM jobs"); got != 2 {
			t.Fatalf("jobs = %d", got)
		}

		_ = opened.Close()
	}
}

func TestVersion8AttemptsKeepTheirMetadataAndGainANullableProfile(t *testing.T) {
	t.Parallel()

	dir := filepath.Join(t.TempDir(), "state")
	rawStore(t, dir, schemaSQL,
		"INSERT INTO attempts (id,job_id,generation,node,started,provider,model)"+
			" VALUES ('a','j',1,'node-a',0,'codex','model-x')",
		"PRAGMA user_version=8",
	)

	for range 2 {
		opened := freshStore(t, dir)

		var (
			provider, model string
			profile         sql.NullString
		)

		err := opened.QueryRowContext(t.Context(), "SELECT provider,model,profile FROM attempts WHERE id='a'").
			Scan(&provider, &model, &profile)
		if err != nil || provider != "codex" || model != "model-x" || profile.Valid {
			t.Fatalf("attempt = %q %q %v, %v", provider, model, profile, err)
		}

		if got := scan[int](t, opened, "PRAGMA user_version"); got != Version {
			t.Fatalf("user_version = %d", got)
		}

		_ = opened.Close()
	}
}

func TestTheHotPathIndexesAndLaterColumnsExist(t *testing.T) {
	t.Parallel()

	opened := freshStore(t, t.TempDir())

	for _, index := range []string{"results_job", "results_feed", "jobs_outstanding", "jobs_payloads", "audit_created"} {
		query := "SELECT count(*) FROM sqlite_master WHERE type='index' AND name='" + index + "'"
		if got := scan[int](t, opened, query); got != 1 {
			t.Fatalf("index %s missing", index)
		}
	}

	_, err := opened.ExecContext(t.Context(), "SELECT exclude_providers, retry_of, pin, tier FROM jobs")
	if err != nil {
		t.Fatal(err)
	}
}
