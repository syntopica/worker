import sqlite3

from tests.conftest import fresh_store
from worker.store.schema_sql import SCHEMA

OLD = (
    SCHEMA.replace(
        "  tokens_out INTEGER,\n  provider TEXT,\n  cost_usd REAL,\n  model TEXT\n",
        "  tokens_out INTEGER\n",
    )
    .replace("  tier TEXT NOT NULL DEFAULT 'basic',\n", "")
    .replace("  acked REAL,\n  rating TEXT\n", "  acked REAL\n")
)


def test_an_older_store_gains_the_ledger_and_quality_columns(tmp_path):
    assert "cost_usd" not in OLD and "tier" not in OLD and "rating" not in OLD
    state = tmp_path / "state"
    state.mkdir()
    conn = sqlite3.connect(state / "meta.sqlite3")
    conn.execute("ATTACH DATABASE ? AS p", (str(state / "payloads.sqlite3"),))
    conn.executescript(OLD)
    conn.execute(
        "INSERT INTO jobs (id, producer, queue, kind, idempotency_key, payload_hash, priority, privacy,"
        " model, state, max_attempts, not_before, created, updated)"
        " VALUES ('old', 'pa', 'pa.bulk', 'inference', 'k', 'h', 50, 'mail', 'model-a', 'queued', 3,"
        " 0, 0, 0)"
    )
    conn.execute("PRAGMA user_version=4")
    conn.commit()
    conn.close()
    store = fresh_store(state)
    columns = {r[1] for r in store.execute("PRAGMA table_info(attempts)")}
    assert {"provider", "cost_usd", "model"} <= columns
    assert "rating" in {r[1] for r in store.execute("PRAGMA table_info(results)")}
    assert store.execute("SELECT tier FROM jobs").fetchone()[0] == "basic"
