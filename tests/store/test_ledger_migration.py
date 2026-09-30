import sqlite3

from tests.conftest import fresh_store
from worker.store.schema_sql import SCHEMA

OLD = SCHEMA.replace(
    "  tokens_out INTEGER,\n  provider TEXT,\n  cost_usd REAL\n", "  tokens_out INTEGER\n"
)


def test_an_older_store_gains_the_ledger_columns(tmp_path):
    state = tmp_path / "state"
    state.mkdir()
    conn = sqlite3.connect(state / "meta.sqlite3")
    conn.execute("ATTACH DATABASE ? AS p", (str(state / "payloads.sqlite3"),))
    conn.executescript(OLD)
    conn.execute("PRAGMA user_version=4")
    conn.close()
    columns = {r[1] for r in fresh_store(state).execute("PRAGMA table_info(attempts)")}
    assert {"provider", "cost_usd"} <= columns
